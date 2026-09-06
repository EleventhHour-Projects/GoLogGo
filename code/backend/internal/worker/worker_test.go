package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parsergen"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// mockPublisher is a thread-safe mock for MessagePublisher
type mockPublisher struct {
	mu       sync.Mutex
	messages [][]byte
}

func (m *mockPublisher) PublishMessage(ctx context.Context, message []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockPublisher) MessageCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func TestNewWorker(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	pub := &mockPublisher{}
	w := NewWorker(nil, nil, pub, jobChan, 5)

	if w == nil {
		t.Fatal("expected non-nil Worker")
	}
	if w.numWorkers != 5 {
		t.Errorf("expected numWorkers 5, got %d", w.numWorkers)
	}
	if w.JobChan != jobChan {
		t.Errorf("expected JobChan to match")
	}
	if w.RMQ != pub {
		t.Errorf("expected RMQ publisher to match")
	}
}

func TestWorker_RunShutdownOnContextCancel(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	w := NewWorker(nil, nil, nil, jobChan, 1)

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
	w := NewWorker(nil, nil, nil, jobChan, 1)

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

func TestWorker_DuplicateFingerprint_SinglePublish_Simulation_Duplicate(t *testing.T) {
	// Simulate 100 concurrent workers encountering the same fingerprint
	// Test that an atomic check/lock ensures only 1 worker publishes the message.
	var publishCount int32
	var pendingLock sync.Map

	pub := &mockPublisher{}
	rawLog := "2026-09-05 18:30:21 ERROR auth-service server-01 Login failed"
	features := fingerprint.ExtractFingerprintFeatures(rawLog)
	hash := fingerprint.GenerateFingerprintHash(features)
	statusKey := fmt.Sprintf("fingerprint:status:%s", hash)

	var wg sync.WaitGroup
	workerCount := 100

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Atomic simulate SetNX: LoadOrStore returns loaded=false if key was newly set
			_, loaded := pendingLock.LoadOrStore(statusKey, "PENDING")
			acquiredPendingLock := !loaded

			if acquiredPendingLock {
				atomic.AddInt32(&publishCount, 1)
				msg := parsergen.ParserGenMessage{
					Hash:     hash,
					RawLog:   rawLog,
					Features: features,
				}
				msgBytes, _ := json.Marshal(msg)
				_ = pub.PublishMessage(context.Background(), msgBytes)
			}
		}(i)
	}

	wg.Wait()

	if publishCount != 1 {
		t.Errorf("expected exactly 1 worker to acquire lock and publish, got %d", publishCount)
	}

	if pub.MessageCount() != 1 {
		t.Errorf("expected exactly 1 published message, got %d", pub.MessageCount())
	}
}
