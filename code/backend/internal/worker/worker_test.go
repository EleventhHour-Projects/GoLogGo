package worker

import (
	"context"
	"testing"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/ml"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewWorker(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	w := NewWorker(nil, nil, jobChan, 5)

	if w == nil {
		t.Fatal("expected non-nil Worker")
	}
	if w.numWorkers != 5 {
		t.Errorf("expected numWorkers 5, got %d", w.numWorkers)
	}
	if w.JobChan != jobChan {
		t.Errorf("expected JobChan to match")
	}
	if w.MLClient == nil {
		t.Error("expected MLClient to be initialized")
	}
}

func TestNewWorkerWithML(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	customML := ml.NewClient("http://mock-ml:8000")
	w := NewWorkerWithML(nil, nil, jobChan, 2, customML)

	if w.MLClient != customML {
		t.Errorf("expected MLClient to match customML instance")
	}
}

func TestWorker_RunShutdownOnContextCancel(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	w := NewWorker(nil, nil, jobChan, 1)

	ctx, cancel := context.WithCancel(context.Background())
	w.wg.Add(1)

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	// Cancel context to signal worker exit
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit in a timely manner on context cancel")
	}
}

func TestWorker_RunShutdownOnChannelClose(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	w := NewWorker(nil, nil, jobChan, 1)

	ctx := context.Background()
	w.wg.Add(1)

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	// Close job channel
	close(jobChan)

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit in a timely manner on channel close")
	}
}
