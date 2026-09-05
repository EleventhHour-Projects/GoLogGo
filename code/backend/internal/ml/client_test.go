package ml

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
)

func TestRequestParser_DummyFallback(t *testing.T) {
	client := NewClient("")
	rawLog := "2026-09-05 18:30:21 ERROR payment-service server-01 Payment failed"
	features := fingerprint.FingerprintFeatures{Format: fingerprint.FormatPlainText}

	p, err := client.RequestParser(context.Background(), rawLog, features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p == nil {
		t.Fatal("expected non-nil parser")
	}

	normalized, err := parser.ParseLog(rawLog, p)
	if err != nil {
		t.Fatalf("failed to parse with dummy parser: %v", err)
	}

	if normalized.Severity != "ERROR" {
		t.Errorf("expected severity ERROR, got %q", normalized.Severity)
	}
	if normalized.Service != "payment-service" {
		t.Errorf("expected service payment-service, got %q", normalized.Service)
	}
	if normalized.Message != "Payment failed" {
		t.Errorf("expected message 'Payment failed', got %q", normalized.Message)
	}
}

func TestRequestParser_FastAPIServer(t *testing.T) {
	expectedParser := parser.Parser{
		Pattern: `^(?P<time>\S+) (?P<msg>.*)$`,
		Mapping: map[string]string{
			"time": "timestamp",
			"msg":  "message",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/parse" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req ParseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedParser)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	rawLog := "2026-09-05T12:00:00Z Test log line"
	features := fingerprint.FingerprintFeatures{Format: fingerprint.FormatPlainText}

	p, err := client.RequestParser(context.Background(), rawLog, features)
	if err != nil {
		t.Fatalf("unexpected error from FastAPI mock: %v", err)
	}

	if p.Pattern != expectedParser.Pattern {
		t.Errorf("expected pattern %q, got %q", expectedParser.Pattern, p.Pattern)
	}
}
