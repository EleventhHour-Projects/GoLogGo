package api

import (
	"bytes"

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

	// Save the raw request to MongoDB.
	logReq, err := api.Reqs.CreateReq(req.HTTPRequest.Context(), body)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to save log request",
		})
	}

	return res.JSON(http.StatusAccepted, map[string]interface{}{
		"status": "accepted",
		"id":     logReq.ID,
	})
}

