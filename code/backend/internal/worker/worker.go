package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/ml"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/redis"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Worker struct {
	Reqs       *database.MongoDB
	Redis      *redis.Redis
	JobChan    chan bson.ObjectID
	numWorkers int
	wg         sync.WaitGroup
	MLClient   *ml.Client
}

// NewWorker instantiates a new Worker struct with required dependencies.
func NewWorker(reqs *database.MongoDB, rdb *redis.Redis, jobChan chan bson.ObjectID, numWorkers int) *Worker {
	mlURL := os.Getenv("ML_API_URL")
	return &Worker{
		Reqs:       reqs,
		Redis:      rdb,
		JobChan:    jobChan,
		numWorkers: numWorkers,
		MLClient:   ml.NewClient(mlURL),
	}
}

// NewWorkerWithML instantiates a Worker with a custom ML client (useful for testing and dependency injection).
func NewWorkerWithML(reqs *database.MongoDB, rdb *redis.Redis, jobChan chan bson.ObjectID, numWorkers int, mlClient *ml.Client) *Worker {
	return &Worker{
		Reqs:       reqs,
		Redis:      rdb,
		JobChan:    jobChan,
		numWorkers: numWorkers,
		MLClient:   mlClient,
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

// processJob coordinates fetching, parsing, caching, and storing logs.
func (w *Worker) processJob(ctx context.Context, jobID bson.ObjectID) error {
	// 1. Get Log Entry from MongoDB requests collection
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

	// 4. Lookup Redis for key: parser:<hash>
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

	// 5. If not found in Redis, call FastAPI ML service (or fallback dummy) to get a new parser
	if p == nil {
		log.Printf("Parser cache miss for key %s, requesting from ML service", redisKey)
		if w.MLClient == nil {
			w.MLClient = ml.NewClient(os.Getenv("ML_API_URL"))
		}

		newParser, mlErr := w.MLClient.RequestParser(ctx, rawLog, features)
		if mlErr != nil {
			_ = w.Reqs.IncrementReqAttempts(ctx, jobID)
			_ = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusFailed)
			return fmt.Errorf("failed to fetch parser from ML service: %w", mlErr)
		}

		p = newParser

		// Save new parser format to Redis under parser:<hash>
		if w.Redis != nil && w.Redis.Client != nil && p != nil {
			if parserJSON, err := json.Marshal(p); err == nil {
				if setErr := w.Redis.Client.Set(ctx, redisKey, parserJSON, 0).Err(); setErr != nil {
					log.Printf("Warning: failed to cache parser in Redis for key %s: %v", redisKey, setErr)
				} else {
					log.Printf("Successfully cached new parser in Redis for key %s", redisKey)
				}
			}
		}
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