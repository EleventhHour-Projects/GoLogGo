package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/auth"
	"github.com/nottechdm/notnet/pkg/notnet"
)

// LogHandler accepts raw log payloads from arbitrary sources.
// It intentionally keeps the parser stubbed for future universal parsing logic.
func (api *Config) LogHandler(req *notnet.Request, res *notnet.Response, claims *auth.CustomClaims) error {
	body, err := io.ReadAll(req.HTTPRequest.Body)
	if err != nil {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "unable to read request body",
		})
	}
	defer req.HTTPRequest.Body.Close()

	if len(bytes.TrimSpace(body)) == 0 {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "empty log payload",
		})
	}

	payload := map[string]interface{}{
		"received":       true,
		"content_type":   req.HTTPRequest.Header.Get("Content-Type"),
		"body_size":      len(body),
		"status":         "accepted",
		"source":         req.RemoteAddr(),
		"request_method": req.Method(),
	}

	// TODO: implement universal parser that accepts logs from network sources
	// and normalizes them into a common internal log structure before storage.
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		payload["format"] = "json"
		payload["parsed"] = parsed
		return res.JSON(http.StatusAccepted, payload)
	}

	payload["format"] = "raw_text"
	payload["raw_log"] = string(body)
	return res.JSON(http.StatusAccepted, payload)
}
