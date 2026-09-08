package parsergen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/ml"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/redis"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ParserGenerator consumes parser-generation jobs from RabbitMQ, requests parsers from the ML service,
// caches parsers in Redis, and requeues waiting logs into the JobChan.
type ParserGenerator struct {
	Reqs     *database.MongoDB
	Redis    *redis.Redis
	RMQ      MessageConsumer
	MLClient MLParserClient
	JobChan  chan bson.ObjectID
}

// New creates a new ParserGenerator with default ML client.
func New(reqs *database.MongoDB, rdb *redis.Redis, rmq MessageConsumer, jobChan chan bson.ObjectID) *ParserGenerator {
	mlURL := os.Getenv("ML_API_URL")
	return &ParserGenerator{
		Reqs:     reqs,
		Redis:    rdb,
		RMQ:      rmq,
		MLClient: ml.NewClient(mlURL),
		JobChan:  jobChan,
	}
}

// NewWithML creates a ParserGenerator with a custom ML client (for testing/dependency injection).
func NewWithML(reqs *database.MongoDB, rdb *redis.Redis, rmq MessageConsumer, jobChan chan bson.ObjectID, mlClient MLParserClient) *ParserGenerator {
	return &ParserGenerator{
		Reqs:     reqs,
		Redis:    rdb,
		RMQ:      rmq,
		MLClient: mlClient,
		JobChan:  jobChan,
	}
}

// Start starts consuming parser-generation messages from RabbitMQ.
func (pg *ParserGenerator) Start(ctx context.Context) error {
	if pg.RMQ == nil {
		return errors.New("rabbitmq consumer is nil")
	}

	log.Println("ParserGenerator consumer starting...")
	return pg.RMQ.ConsumeMessages(ctx, func(message []byte) error {
		return pg.ProcessMessage(ctx, message)
	})
}

// ProcessMessage processes a single parser generation message.
func (pg *ParserGenerator) ProcessMessage(ctx context.Context, body []byte) error {
	var msg ParserGenMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("ParserGenerator: invalid message json: %v", err)
		return nil // Ack invalid messages so they are not endlessly requeued
	}

	if msg.Hash == "" {
		log.Println("ParserGenerator: message missing hash")
		return nil
	}

	redisKey := fmt.Sprintf("parser:%s", msg.Hash)
	statusKey := fmt.Sprintf("fingerprint:status:%s", msg.Hash)

	var p *parser.Parser

	// 1. Check parser again in Redis
	if pg.Redis != nil && pg.Redis.Client != nil {
		val, err := pg.Redis.Client.Get(ctx, redisKey).Result()
		if err == nil {
			var cached parser.Parser
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				p = &cached
				log.Printf("ParserGenerator: parser already exists for hash %s (cache hit)", msg.Hash)
			}
		}
	}

	// 2. If parser exists: requeue pending logs
	if p != nil {
		if pg.Redis != nil && pg.Redis.Client != nil {
			_ = pg.Redis.Client.Set(ctx, statusKey, "READY", 24*time.Hour).Err()
		}
		pg.requeuePendingLogs(ctx, msg.Hash)
		return nil
	}

	// 3. If parser doesn't exist: call existing ML client
	if pg.MLClient == nil {
		pg.MLClient = ml.NewClient(os.Getenv("ML_API_URL"))
	}

	newParser, err := pg.MLClient.RequestParser(ctx, msg.RawLog, msg.Features)
	if err != nil {
		log.Printf("ParserGenerator: failed to generate parser for hash %s: %v", msg.Hash, err)
		// Ensure fingerprint is NOT permanently stuck in PENDING on failure
		if pg.Redis != nil && pg.Redis.Client != nil {
			pg.Redis.Client.Del(ctx, statusKey)
		}
		pg.failPendingLogs(ctx, msg.Hash)
		return fmt.Errorf("failed to generate parser from ML service: %w", err)
	}

	// 4. Save parser to MongoDB & Redis/storage
	if newParser != nil && pg.Reqs != nil {
		formatName := "Custom"
		if msg.Features.Format != "" && msg.Features.Format != "unknown" {
			formatName = string(msg.Features.Format)
		}

		var signatures []map[string]string
		for _, sig := range msg.Features.FieldSignatures {
			signatures = append(signatures, map[string]string{
				"key":  sig.Key,
				"type": string(sig.Type),
			})
		}

		var extracted []string
		if len(msg.Features.Keys) > 0 {
			extracted = msg.Features.Keys
		} else {
			for k := range newParser.Mapping {
				extracted = append(extracted, k)
			}
		}

		doc := database.ParserDoc{
			Hash:            msg.Hash,
			Pattern:         newParser.Pattern,
			Mapping:         newParser.Mapping,
			Transformations: newParser.Transformations,
			Status:          "Active",
			Format:          formatName,
			Vendor:          msg.Features.VendorHint,
			Product:         msg.Features.ProductHint,
			Template:        msg.Features.Template,
			Delimiter:       msg.Features.Delimiter,
			ExtractedFields: extracted,
			FieldSignatures: signatures,
			SampleLog:       msg.RawLog,
		}
		if _, dbErr := pg.Reqs.SaveParser(ctx, &doc); dbErr != nil {
			log.Printf("ParserGenerator: warning: failed to save parser to MongoDB for hash %s: %v", msg.Hash, dbErr)
		} else {
			log.Printf("ParserGenerator: successfully saved parser to MongoDB for hash %s", msg.Hash)
		}
	}

	if pg.Redis != nil && pg.Redis.Client != nil && newParser != nil {
		parserJSON, err := json.Marshal(newParser)
		if err == nil {
			if setErr := pg.Redis.Client.Set(ctx, redisKey, parserJSON, 0).Err(); setErr != nil {
				log.Printf("ParserGenerator: warning: failed to cache parser in Redis for key %s: %v", redisKey, setErr)
			} else {
				log.Printf("ParserGenerator: successfully cached new parser in Redis for key %s", redisKey)
			}
		}

		// 5. Mark fingerprint as READY
		pg.Redis.Client.Set(ctx, statusKey, "READY", 24*time.Hour)
	}

	// 6. Requeue pending logs into the existing JobChan
	pg.requeuePendingLogs(ctx, msg.Hash)

	return nil
}

// requeuePendingLogs fetches all waiting jobs for this fingerprint, updates their MongoDB status to pending,
// and pushes them into the JobChan for processing.
func (pg *ParserGenerator) requeuePendingLogs(ctx context.Context, hash string) {
	if pg.Redis == nil || pg.Redis.Client == nil {
		return
	}

	pendingKey := fmt.Sprintf("fingerprint:pending_jobs:%s", hash)
	jobIDs, err := pg.Redis.Client.SMembers(ctx, pendingKey).Result()
	if err != nil {
		log.Printf("ParserGenerator: failed to fetch pending jobs for hash %s: %v", hash, err)
		return
	}

	for _, idStr := range jobIDs {
		objID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			log.Printf("ParserGenerator: invalid ObjectID in pending jobs: %s", idStr)
			continue
		}

		// Update MongoDB request status to StatusPending so it is ready for processing
		if pg.Reqs != nil {
			_ = pg.Reqs.UpdateReqStatus(ctx, objID, database.StatusPending)
		}

		// Requeue into JobChan
		if pg.JobChan != nil {
			select {
			case pg.JobChan <- objID:
				log.Printf("ParserGenerator: requeued job %s into JobChan", idStr)
			case <-ctx.Done():
				return
			default:
				// If channel is full, status in MongoDB is already StatusPending,
				// so the scheduler will pick it up on next cycle.
				log.Printf("ParserGenerator: JobChan full, job %s queued in MongoDB as pending", idStr)
			}
		}
	}

	// Delete pending jobs set after requeuing
	pg.Redis.Client.Del(ctx, pendingKey)
}

// failPendingLogs updates all waiting requests for a failed fingerprint to StatusFailed.
func (pg *ParserGenerator) failPendingLogs(ctx context.Context, hash string) {
	if pg.Redis == nil || pg.Redis.Client == nil {
		return
	}

	pendingKey := fmt.Sprintf("fingerprint:pending_jobs:%s", hash)
	jobIDs, err := pg.Redis.Client.SMembers(ctx, pendingKey).Result()
	if err != nil {
		return
	}

	for _, idStr := range jobIDs {
		objID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		if pg.Reqs != nil {
			_ = pg.Reqs.IncrementReqAttempts(ctx, objID)
			_ = pg.Reqs.UpdateReqStatus(ctx, objID, database.StatusFailed)
		}
	}

	pg.Redis.Client.Del(ctx, pendingKey)
}
