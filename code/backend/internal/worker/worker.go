package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Worker struct {
	Reqs    *database.MongoDB // save whole MongoDB struct
	JobChan chan bson.ObjectID
	numWorkers int
	wg sync.WaitGroup
}

// NewWorker Instantiates a new Worker Struct  with required data
func NewWorker(reqs *database.MongoDB, JobChan chan bson.ObjectID, numWorkers int) *Worker {
	return &Worker{
		Reqs: reqs,
		JobChan: JobChan,
		numWorkers: numWorkers,
	}
}

func(w *Worker) InitialiseWorkerPool(ctx context.Context) {
	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		go w.Run(ctx)
	}

	<-ctx.Done() // industry standard (not necessary here)
	w.wg.Wait()
}

func(w *Worker) Run(ctx context.Context) {
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
func(w *Worker) processJob(ctx context.Context, jobID bson.ObjectID) error {
	time.Sleep(5 * time.Second)
	log.Printf("Parsed Successfully for ID: %v", jobID)
	return nil
}