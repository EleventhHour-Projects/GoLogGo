package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/fingerprint"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
)

// Client handles interaction with the ML FastAPI parser generator service.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// ParseRequest is the payload sent to the FastAPI ML service.
type ParseRequest struct {
	Log      string                          `json:"log"`
	Features fingerprint.FingerprintFeatures `json:"features"`
}

// NewClient creates a new ML Client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RequestParser requests a new Parser format from the ML service.
// If the FastAPI service is not reachable or not yet implemented, it falls back to a smart dummy parser generator.
func (c *Client) RequestParser(ctx context.Context, rawLog string, features fingerprint.FingerprintFeatures) (*parser.Parser, error) {
	if c != nil && c.BaseURL != "" {
		p, err := c.callFastAPI(ctx, rawLog, features)
		if err == nil && p != nil {
			return p, nil
		}
	}

	// Fallback dummy parser when FastAPI is unavailable
	return GenerateDummyParser(rawLog, features), nil
}

// callFastAPI performs the HTTP POST request to FastAPI.
func (c *Client) callFastAPI(ctx context.Context, rawLog string, features fingerprint.FingerprintFeatures) (*parser.Parser, error) {
	reqBody, err := json.Marshal(ParseRequest{
		Log:      rawLog,
		Features: features,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parse request: %w", err)
	}

	url := fmt.Sprintf("%s/parse", strings.TrimRight(c.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request to ml service failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml service returned non-200 status: %d", resp.StatusCode)
	}

	var p parser.Parser
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode parser from ml service response: %w", err)
	}

	return &p, nil
}

// GenerateDummyParser generates a sensible parser for logs when ML service is mocked/dummy.
func GenerateDummyParser(rawLog string, features fingerprint.FingerprintFeatures) *parser.Parser {
	trimmed := strings.TrimSpace(rawLog)

	// Pattern 1: Common 5-part format: timestamp severity service host message
	// Example: 2026-09-05 18:30:21 ERROR payment-service server-01 Payment failed
	pattern5 := `^(?P<timestamp>\S+ \S+) (?P<severity>[A-Za-z]+) (?P<service>\S+) (?P<host>\S+) (?P<message>.*)$`
	if re, err := regexp.Compile(pattern5); err == nil && re.MatchString(trimmed) {
		return &parser.Parser{
			Pattern: pattern5,
			Mapping: map[string]string{
				"timestamp": "timestamp",
				"severity":  "severity",
				"service":   "service",
				"host":      "host",
				"message":   "message",
			},
			Transformations: map[string]parser.Transformation{
				"severity": {
					Type: "uppercase",
				},
				"timestamp": {
					Type:   "datetime",
					Format: "2006-01-02 15:04:05",
				},
			},
		}
	}

	// Pattern 2: Common ISO timestamp + severity + service + message
	// Example: 2026-09-05T18:30:21Z [INFO] [auth-service] User logged in
	pattern4 := `^(?P<timestamp>\S+) \[(?P<severity>[A-Za-z]+)\] \[(?P<service>\S+)\] (?P<message>.*)$`
	if re, err := regexp.Compile(pattern4); err == nil && re.MatchString(trimmed) {
		return &parser.Parser{
			Pattern: pattern4,
			Mapping: map[string]string{
				"timestamp": "timestamp",
				"severity":  "severity",
				"service":   "service",
				"message":   "message",
			},
			Transformations: map[string]parser.Transformation{
				"severity": {
					Type: "uppercase",
				},
				"timestamp": {
					Type: "datetime",
				},
			},
		}
	}

	// Pattern 3: Simple timestamp + severity + message
	// Example: 2026-09-05 18:30:21 INFO Something happened
	pattern3 := `^(?P<timestamp>\S+ \S+) (?P<severity>[A-Za-z]+) (?P<message>.*)$`
	if re, err := regexp.Compile(pattern3); err == nil && re.MatchString(trimmed) {
		return &parser.Parser{
			Pattern: pattern3,
			Mapping: map[string]string{
				"timestamp": "timestamp",
				"severity":  "severity",
				"message":   "message",
			},
			Transformations: map[string]parser.Transformation{
				"severity": {
					Type: "uppercase",
				},
				"timestamp": {
					Type:   "datetime",
					Format: "2006-01-02 15:04:05",
				},
			},
		}
	}

	// Pattern 4: Universal fallback capturing the entire log as message
	return &parser.Parser{
		Pattern: `^(?P<message>.*)$`,
		Mapping: map[string]string{
			"message": "message",
		},
	}
}
