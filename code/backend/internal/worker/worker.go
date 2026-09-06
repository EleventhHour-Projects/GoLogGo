package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parsergen"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/redis"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Worker struct {
	Reqs       *database.MongoDB
	Redis      *redis.Redis
	RMQ        MessagePublisher
	JobChan    chan bson.ObjectID
	numWorkers int
	wg         sync.WaitGroup
}

// NewWorker instantiates a new Worker struct with required dependencies.
func NewWorker(reqs *database.MongoDB, rdb *redis.Redis, rmq MessagePublisher, jobChan chan bson.ObjectID, numWorkers int) *Worker {
	return &Worker{
		Reqs:       reqs,
		Redis:      rdb,
		RMQ:        rmq,
		JobChan:    jobChan,
		numWorkers: numWorkers,
	}
}

func (w *Worker) InitialiseWorkerPool(ctx context.Context) {
	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		go w.Run(ctx)
	}

	<-ctx.Done()
	w.wg.Wait()
}

func (w *Worker) Run(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case jobID, ok := <-w.JobChan:
			if !ok {
				log.Println("Channel closed, exiting")
				return
			}
			err := w.processJob(ctx, jobID)
			if err != nil {
				log.Println("Error processing job:", err)
			}
		}
	}
}

// processJob coordinates fetching, fingerprinting, parser checking, parsing, and storing logs.
func (w *Worker) processJob(ctx context.Context, jobID bson.ObjectID) error {
	// 1. Fetch request from MongoDB requests collection
	req, err := w.Reqs.FindReqByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to find request %v: %w", jobID, err)
	}

	// 2. Update the status of the request to "processing"
	err = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusProcessing)
	if err != nil {
		return fmt.Errorf("failed to update status to processing for %v: %w", jobID, err)
	}

	rawLog := string(req.Payload)

	// 3. Extract Fingerprint Features and Hash from Log Entry
	features := fingerprint.ExtractFingerprintFeatures(rawLog)
	hash := fingerprint.GenerateFingerprintHash(features)
	redisKey := fmt.Sprintf("parser:%s", hash)

	var p *parser.Parser

	// 4. Check Redis for key: parser:<hash>
	if w.Redis != nil && w.Redis.Client != nil {
		val, err := w.Redis.Client.Get(ctx, redisKey).Result()
		if err == nil {
			var cached parser.Parser
			if err := json.Unmarshal([]byte(val), &cached); err == nil {
				p = &cached
				log.Printf("Parser cache hit in Redis for key %s", redisKey)
			}
		}
	}

	// 5. If parser doesn't exist:
	//    - mark fingerprint as PENDING atomically
	//    - only the worker that successfully creates PENDING publishes a RabbitMQ parser-generation job
	//    - mark request as waiting for parser
	//    - DO NOT call ML
	if p == nil {
		log.Printf("Parser not found for hash %s, queuing for parser generation", hash)

		statusKey := fmt.Sprintf("fingerprint:status:%s", hash)
		pendingJobsKey := fmt.Sprintf("fingerprint:pending_jobs:%s", hash)

		var acquiredPendingLock bool
		if w.Redis != nil && w.Redis.Client != nil {
			// Atomically set fingerprint status to PENDING with a 5-minute TTL
			ok, setErr := w.Redis.Client.SetNX(ctx, statusKey, "PENDING", 5*time.Minute).Result()
			if setErr != nil {
				log.Printf("Warning: failed to set atomic pending status in Redis for %s: %v", hash, setErr)
			} else {
				acquiredPendingLock = ok
			}

			// Track this request as waiting for the parser
			if err := w.Redis.Client.SAdd(ctx, pendingJobsKey, jobID.Hex()).Err(); err != nil {
				log.Printf("Warning: failed to add job %v to pending jobs set: %v", jobID, err)
			}
			_ = w.Redis.Client.Expire(ctx, pendingJobsKey, 10*time.Minute)
		} else {
			// Fallback if redis is nil (e.g. testing without redis)
			acquiredPendingLock = true
		}

		// Only the worker that successfully created PENDING publishes the RabbitMQ job
		if acquiredPendingLock && w.RMQ != nil {
			msg := parsergen.ParserGenMessage{
				Hash:     hash,
				RawLog:   rawLog,
				Features: features,
			}
			msgBytes, jsonErr := json.Marshal(msg)
			if jsonErr == nil {
				if pubErr := w.RMQ.PublishMessage(ctx, msgBytes); pubErr != nil {
					log.Printf("Failed to publish parser-generation job to RabbitMQ for hash %s: %v", hash, pubErr)
					if w.Redis != nil && w.Redis.Client != nil {
						w.Redis.Client.Del(ctx, statusKey)
					}
					_ = w.Reqs.IncrementReqAttempts(ctx, jobID)
					_ = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusFailed)
					return fmt.Errorf("failed to publish parser-generation job: %w", pubErr)
				}
				log.Printf("Successfully published parser-generation job to RabbitMQ for hash %s", hash)
			}
		}

		// Mark request as waiting for parser
		if err := w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusWaitingParser); err != nil {
			return fmt.Errorf("failed to update request status to waiting_parser for %v: %w", jobID, err)
		}

		return nil
	}

	// 6. Normalize the raw log using the parser engine
	normalizedLog, err := parser.ParseLog(rawLog, p)
	if err != nil {
		_ = w.Reqs.IncrementReqAttempts(ctx, jobID)
		_ = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusFailed)
		return fmt.Errorf("failed to normalize log for %v (hash: %s): %w", jobID, hash, err)
	}

	// 7. Insert the normalized log and original raw log into the MongoDB logs collection
	err = w.Reqs.InsertLog(ctx, *req, *normalizedLog, false, req.UserID, hash)
	if err != nil {
		_ = w.Reqs.IncrementReqAttempts(ctx, jobID)
		_ = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusFailed)
		return fmt.Errorf("failed to insert normalized log to database for %v: %w", jobID, err)
	}

	// 8. Update request status to "completed"
	err = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to update request status to completed for %v: %w", jobID, err)
	}

	log.Printf("Successfully processed and saved log for JobID: %v (Hash: %s)", jobID, hash)
	return nil
}
