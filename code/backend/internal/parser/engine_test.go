package parser

import (
	"reflect"
	"testing"
)

func TestParseLog_BasicExample(t *testing.T) {
	rawLog := "2026-09-05 18:30:21 ERROR payment-service server-01 Payment failed"

	p := &Parser{
		Pattern: `^(?P<timestamp>\S+ \S+) (?P<severity>\w+) (?P<service>\S+) (?P<host>\S+) (?P<message>.*)$`,
		Mapping: map[string]string{
			"timestamp": "timestamp",
			"severity":  "severity",
			"service":   "service",
			"host":      "host",
			"message":   "message",
		},
		Transformations: map[string]Transformation{
			"severity": {
				Type: "uppercase",
			},
			"timestamp": {
				Type:   "datetime",
				Format: "2006-01-02 15:04:05",
			},
		},
	}

	normalized, err := ParseLog(rawLog, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if normalized.Timestamp != "2026-09-05T18:30:21Z" {
		t.Errorf("expected timestamp '2026-09-05T18:30:21Z', got %q", normalized.Timestamp)
	}
	if normalized.Severity != "ERROR" {
		t.Errorf("expected severity 'ERROR', got %q", normalized.Severity)
	}
	if normalized.Service != "payment-service" {
		t.Errorf("expected service 'payment-service', got %q", normalized.Service)
	}
	if normalized.Host != "server-01" {
		t.Errorf("expected host 'server-01', got %q", normalized.Host)
	}
	if normalized.Message != "Payment failed" {
		t.Errorf("expected message 'Payment failed', got %q", normalized.Message)
	}
	if len(normalized.Metadata) != 0 {
		t.Errorf("expected empty metadata, got %+v", normalized.Metadata)
	}
}

func TestParseLog_RejectsTrailingInput(t *testing.T) {
	p := &Parser{Pattern: `^(?P<severity>\w+) (?P<message>.*)$`, Mapping: map[string]string{
		"severity": "severity",
		"message":  "message",
	}}

	if _, err := ParseLog("INFO message trailing", p); err != nil {
		t.Fatalf("expected a valid full-line match, got %v", err)
	}

	anchoredParser := &Parser{Pattern: `^INFO$`, Mapping: map[string]string{}}
	if _, err := ParseLog("INFO unexpected", anchoredParser); err == nil {
		t.Fatal("expected trailing input to be rejected")
	}
}

func TestParseLog_ExtractedFieldTransformationKey(t *testing.T) {
	// Raw log with lowercase severity and raw_time field
	rawLog := "raw_time=2026-09-05 18:30:21 raw_sev=warn msg=something happened"

	p := &Parser{
		Pattern: `^raw_time=(?P<raw_time>[^ ]+) (?P<raw_time_rest>[^ ]+) raw_sev=(?P<raw_sev>\w+) msg=(?P<raw_msg>.*)$`,
		Mapping: map[string]string{
			"raw_sev": "severity",
			"raw_msg": "message",
		},
		Transformations: map[string]Transformation{
			"raw_sev": {
				Type: "uppercase",
			},
		},
	}

	normalized, err := ParseLog(rawLog, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if normalized.Severity != "WARN" {
		t.Errorf("expected severity 'WARN', got %q", normalized.Severity)
	}
	if normalized.Message != "something happened" {
		t.Errorf("expected message 'something happened', got %q", normalized.Message)
	}
}

func TestParseLog_MetadataFields(t *testing.T) {
	rawLog := "2026-09-05T12:00:00Z INFO user-service host-a user_id=12345 ip=192.168.1.100 action=login"

	p := &Parser{
		Pattern: `^(?P<time>\S+) (?P<level>\w+) (?P<app>\S+) (?P<host>\S+) user_id=(?P<uid>\d+) ip=(?P<ip>\S+) action=(?P<action>\w+)$`,
		Mapping: map[string]string{
			"time":   "timestamp",
			"level":  "severity",
			"app":    "service",
			"host":   "host",
			"uid":    "metadata.user_id",
			"ip":     "metadata.ip_address",
			"action": "action", // non-standard field without metadata. prefix
		},
		Transformations: map[string]Transformation{
			"level": {Type: "lowercase"},
		},
	}

	normalized, err := ParseLog(rawLog, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if normalized.Severity != "info" {
		t.Errorf("expected severity 'info', got %q", normalized.Severity)
	}
	if normalized.Timestamp != "2026-09-05T12:00:00Z" {
		t.Errorf("expected timestamp '2026-09-05T12:00:00Z', got %q", normalized.Timestamp)
	}
	if normalized.Service != "user-service" {
		t.Errorf("expected service 'user-service', got %q", normalized.Service)
	}
	if normalized.Metadata["user_id"] != "12345" {
		t.Errorf("expected metadata.user_id '12345', got %v", normalized.Metadata["user_id"])
	}
	if normalized.Metadata["ip_address"] != "192.168.1.100" {
		t.Errorf("expected metadata.ip_address '192.168.1.100', got %v", normalized.Metadata["ip_address"])
	}
	if normalized.Metadata["action"] != "login" {
		t.Errorf("expected metadata.action 'login', got %v", normalized.Metadata["action"])
	}
}

func TestParseLog_AllStandardFields(t *testing.T) {
	rawLog := "2026-09-05 10:00:00 WARN auth auth-srv auth_src trace-123 req-456 user_login Login attempt failed"

	p := &Parser{
		Pattern: `^(?P<ts>\S+ \S+) (?P<sev>\w+) (?P<svc>\S+) (?P<hst>\S+) (?P<src>\S+) (?P<trace>\S+) (?P<req>\S+) (?P<evt>\S+) (?P<msg>.*)$`,
		Mapping: map[string]string{
			"ts":    "timestamp",
			"sev":   "severity",
			"svc":   "service",
			"hst":   "host",
			"src":   "source",
			"trace": "trace_id",
			"req":   "request_id",
			"evt":   "event_type",
			"msg":   "message",
		},
		Transformations: map[string]Transformation{
			"ts": {
				Type:   "datetime",
				Format: "YYYY-MM-DD HH:mm:ss", // test strftime-style format conversion
			},
		},
	}

	normalized, err := p.Parse(rawLog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &NormalizedLog{
		Timestamp: "2026-09-05T10:00:00Z",
		Severity:  "WARN",
		Service:   "auth",
		Host:      "auth-srv",
		Source:    "auth_src",
		Message:   "Login attempt failed",
		EventType: "user_login",
		TraceID:   "trace-123",
		RequestID: "req-456",
		Metadata:  map[string]interface{}{},
	}

	if !reflect.DeepEqual(normalized, expected) {
		t.Errorf("normalized log mismatch.\nGot:  %+v\nWant: %+v", normalized, expected)
	}
}

func TestParseLog_Transformations(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		trans       Transformation
		expected    string
		expectError bool
	}{
		{
			name:        "uppercase",
			input:       "error",
			trans:       Transformation{Type: "uppercase"},
			expected:    "ERROR",
			expectError: false,
		},
		{
			name:        "lowercase",
			input:       "WARNING",
			trans:       Transformation{Type: "lowercase"},
			expected:    "warning",
			expectError: false,
		},
		{
			name:        "trim",
			input:       "  padded message  ",
			trans:       Transformation{Type: "trim"},
			expected:    "padded message",
			expectError: false,
		},
		{
			name:        "integer",
			input:       "001204",
			trans:       Transformation{Type: "integer"},
			expected:    "1204",
			expectError: false,
		},
		{
			name:        "float",
			input:       "249.990",
			trans:       Transformation{Type: "float"},
			expected:    "249.99",
			expectError: false,
		},
		{
			name:        "datetime with Go layout",
			input:       "2026-09-05 18:30:21",
			trans:       Transformation{Type: "datetime", Format: "2006-01-02 15:04:05"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "datetime with ISO layout tokens",
			input:       "2026-09-05 18:30:21",
			trans:       Transformation{Type: "datetime", Format: "YYYY-MM-DD HH:mm:ss"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "datetime auto-detect RFC3339",
			input:       "2026-09-05T18:30:21Z",
			trans:       Transformation{Type: "datetime"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "datetime auto-detect Apache/Nginx CLF",
			input:       "05/Sep/2026:18:30:21 +0000",
			trans:       Transformation{Type: "datetime"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "epoch seconds",
			input:       "1788633021",
			trans:       Transformation{Type: "epoch_s"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "epoch milliseconds",
			input:       "1788633021000",
			trans:       Transformation{Type: "epoch_ms"},
			expected:    "2026-09-05T18:30:21Z",
			expectError: false,
		},
		{
			name:        "invalid datetime format",
			input:       "not-a-date",
			trans:       Transformation{Type: "datetime", Format: "2006-01-02"},
			expectError: true,
		},
		{
			name:        "unknown transformation",
			input:       "hello",
			trans:       Transformation{Type: "invalid_type"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := applyTransformation(tt.input, tt.trans)
			if (err != nil) != tt.expectError {
				t.Fatalf("applyTransformation() error = %v, expectError = %v", err, tt.expectError)
			}
			if !tt.expectError && res != tt.expected {
				t.Errorf("applyTransformation() = %q, want %q", res, tt.expected)
			}
		})
	}
}

func TestParseLog_ErrorCases(t *testing.T) {
	t.Run("nil parser", func(t *testing.T) {
		_, err := ParseLog("test", nil)
		if err == nil {
			t.Errorf("expected error for nil parser, got nil")
		}
	})

	t.Run("empty pattern", func(t *testing.T) {
		p := &Parser{Pattern: ""}
		_, err := ParseLog("test", p)
		if err == nil {
			t.Errorf("expected error for empty pattern, got nil")
		}
	})

	t.Run("invalid regex pattern", func(t *testing.T) {
		p := &Parser{Pattern: "[invalid("}
		_, err := ParseLog("test", p)
		if err == nil {
			t.Errorf("expected error for invalid regex, got nil")
		}
	})

	t.Run("log does not match pattern", func(t *testing.T) {
		p := &Parser{Pattern: `^exact_match$`}
		_, err := ParseLog("different text", p)
		if err == nil {
			t.Errorf("expected error when log doesn't match, got nil")
		}
	})

	t.Run("mapped field not found in extracted fields", func(t *testing.T) {
		p := &Parser{
			Pattern: `^(?P<msg>.*)$`,
			Mapping: map[string]string{
				"nonexistent_field": "message",
			},
		}
		_, err := ParseLog("hello world", p)
		if err == nil {
			t.Errorf("expected error for nonexistent mapped field, got nil")
		}
	})

	t.Run("failed transformation", func(t *testing.T) {
		p := &Parser{
			Pattern: `^(?P<ts>.*)$`,
			Mapping: map[string]string{
				"ts": "timestamp",
			},
			Transformations: map[string]Transformation{
				"ts": {
					Type:   "datetime",
					Format: "2006-01-02",
				},
			},
		}
		_, err := ParseLog("invalid-date-string", p)
		if err == nil {
			t.Errorf("expected error for failed transformation, got nil")
		}
	})
}

func TestParseLog_NoExplicitMapping(t *testing.T) {
	rawLog := "2026-09-05 18:30:21 server-01 hello"
	p := &Parser{
		Pattern: `^(?P<timestamp>\S+ \S+) (?P<host>\S+) (?P<message>.*)$`,
		Transformations: map[string]Transformation{
			"timestamp": {
				Type:   "datetime",
				Format: "2006-01-02 15:04:05",
			},
		},
	}

	normalized, err := ParseLog(rawLog, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if normalized.Timestamp != "2026-09-05T18:30:21Z" {
		t.Errorf("expected timestamp '2026-09-05T18:30:21Z', got %q", normalized.Timestamp)
	}
	if normalized.Host != "server-01" {
		t.Errorf("expected host 'server-01', got %q", normalized.Host)
	}
	if normalized.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", normalized.Message)
	}
}
