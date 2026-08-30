package worker

import (
	"context"
	"log"
	"sync"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Worker struct {
	Reqs       *database.MongoDB // save whole MongoDB struct
	JobChan    chan bson.ObjectID
	numWorkers int
	wg         sync.WaitGroup
}

// NewWorker Instantiates a new Worker Struct  with required data
func NewWorker(reqs *database.MongoDB, JobChan chan bson.ObjectID, numWorkers int) *Worker {
	return &Worker{
		Reqs:       reqs,
		JobChan:    JobChan,
		numWorkers: numWorkers,
	}
}

func (w *Worker) InitialiseWorkerPool(ctx context.Context) {
	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		go w.Run(ctx)
	}

	<-ctx.Done() // industry standard (not necessary here)
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

// this is where the job get's processed
func (w *Worker) processJob(ctx context.Context, jobID bson.ObjectID) error {
	// Get Log Entry from MongoDB
	req, err := w.Reqs.FindReqByID(ctx, jobID)
	if err != nil {
		return err
	}
	// Update the status of the request to "processing"
	err = w.Reqs.UpdateReqStatus(ctx, jobID, database.StatusProcessing)
	if err != nil {
		return err
	}

	// Extract Fingerprint Features from Log Entry
	features := fingerprint.ExtractFingerprintFeatures(string(req.Payload))

	// generate fingerprint hash from features
	hash := fingerprint.GenerateFingerprintHash(features)

	log.Printf("Parsed Successfully for ID: %v", jobID)
	log.Printf("Fingerprint Features Field Types: %v", features.Format)
	log.Printf("Fingerprint Hash: %s", hash)
	return nil
}
