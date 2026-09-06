package parsergen

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/ml"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockConsumer struct {
	handler func(message []byte) error
}

func (m *mockConsumer) ConsumeMessages(ctx context.Context, handler func(message []byte) error) error {
	m.handler = handler
	return nil
}

func TestParserGenerator_Initialization(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	consumer := &mockConsumer{}
	pg := New(nil, nil, consumer, jobChan)

	if pg == nil {
		t.Fatal("expected non-nil ParserGenerator")
	}
	if pg.JobChan != jobChan {
		t.Errorf("expected JobChan to match")
	}
	if pg.RMQ != consumer {
		t.Errorf("expected RMQ to match")
	}
	if pg.MLClient == nil {
		t.Errorf("expected MLClient to be initialized")
	}
}

func TestParserGenerator_ProcessMessage_SuccessFlow(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	customML := ml.NewClient("") // Fallback dummy generator
	pg := NewWithML(nil, nil, nil, jobChan, customML)

	rawLog := "2026-09-05 18:30:21 ERROR payment-service server-01 Payment failed"
	features := fingerprint.ExtractFingerprintFeatures(rawLog)
	hash := fingerprint.GenerateFingerprintHash(features)

	msg := ParserGenMessage{
		Hash:     hash,
		RawLog:   rawLog,
		Features: features,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	ctx := context.Background()
	err = pg.ProcessMessage(ctx, body)
	if err != nil {
		t.Fatalf("ProcessMessage returned error: %v", err)
	}
}

func TestParserGenerator_ProcessMessage_CorruptedJSON(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	pg := NewWithML(nil, nil, nil, jobChan, nil)

	invalidJSON := []byte("{invalid-json}")
	err := pg.ProcessMessage(context.Background(), invalidJSON)
	if err != nil {
		t.Errorf("expected nil error for corrupted json (acked), got %v", err)
	}
}

func TestParserGenerator_ProcessMessage_EmptyHash(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	pg := NewWithML(nil, nil, nil, jobChan, nil)

	emptyHashMsg := []byte(`{"hash":""}`)
	err := pg.ProcessMessage(context.Background(), emptyHashMsg)
	if err != nil {
		t.Errorf("expected nil error for empty hash (ignored), got %v", err)
	}
}

type failingMLClient struct{}

func (f *failingMLClient) RequestParser(ctx context.Context, rawLog string, features fingerprint.FingerprintFeatures) (*parser.Parser, error) {
	return nil, fmt.Errorf("simulated ml service outage")
}

func TestParserGenerator_ProcessMessage_MLErrorReturnsError(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	pg := NewWithML(nil, nil, nil, jobChan, &failingMLClient{})

	rawLog := "2026-09-05 18:30:21 ERROR service server-1 failed"
	features := fingerprint.ExtractFingerprintFeatures(rawLog)
	hash := fingerprint.GenerateFingerprintHash(features)

	msg := ParserGenMessage{
		Hash:     hash,
		RawLog:   rawLog,
		Features: features,
	}
	body, _ := json.Marshal(msg)

	err := pg.ProcessMessage(context.Background(), body)
	if err == nil {
		t.Error("expected error from ProcessMessage when ML client fails, got nil")
	}
}

func TestParserGenerator_IdempotentOnDuplicateMessages(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 10)
	customML := ml.NewClient("")
	pg := NewWithML(nil, nil, nil, jobChan, customML)

	rawLog := "2026-09-05 18:30:21 INFO gateway server-01 Request handled"
	features := fingerprint.ExtractFingerprintFeatures(rawLog)
	hash := fingerprint.GenerateFingerprintHash(features)

	msg := ParserGenMessage{
		Hash:     hash,
		RawLog:   rawLog,
		Features: features,
	}
	body, _ := json.Marshal(msg)

	ctx := context.Background()

	// First delivery
	err := pg.ProcessMessage(ctx, body)
	if err != nil {
		t.Fatalf("first ProcessMessage failed: %v", err)
	}

	// Duplicate delivery (should be safe and idempotent)
	err = pg.ProcessMessage(ctx, body)
	if err != nil {
		t.Fatalf("duplicate ProcessMessage failed: %v", err)
	}
}

func TestParserGenerator_StartWithNilRMQ(t *testing.T) {
	pg := New(nil, nil, nil, nil)
	err := pg.Start(context.Background())
	if err == nil {
		t.Errorf("expected error when starting with nil RMQ, got nil")
	}
}

func TestParserGenerator_RequeuePendingLogs(t *testing.T) {
	jobChan := make(chan bson.ObjectID, 5)
	pg := &ParserGenerator{
		JobChan: jobChan,
	}

	// Testing nil redis safe handling
	pg.requeuePendingLogs(context.Background(), "somehash")
	pg.failPendingLogs(context.Background(), "somehash")

	select {
	case <-jobChan:
		t.Fatal("expected no jobs in channel with nil redis")
	default:
		// success
	}
}

// Ensure mockConsumer implements MessageConsumer
var _ MessageConsumer = (*mockConsumer)(nil)
