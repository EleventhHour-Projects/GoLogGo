package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"github.com/nottechdm/notnet/pkg/notnet"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ListParsersHandler returns all parsers from MongoDB along with calculated aggregate metrics.
func (api *Config) ListParsersHandler(req *notnet.Request, res *notnet.Response) error {
	ctx := req.HTTPRequest.Context()
	parsers, err := api.Reqs.FindAllParsers(ctx)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch parsers from database",
		})
	}

	var activeCount int
	var totalLogsProcessed int64
	for _, p := range parsers {
		if strings.EqualFold(p.Status, "active") {
			activeCount++
		}
		totalLogsProcessed += p.LogsProcessed
	}

	return res.JSON(http.StatusOK, map[string]interface{}{
		"parsers": parsers,
		"metrics": map[string]interface{}{
			"totalParsers":       len(parsers),
			"activeParsers":      activeCount,
			"totalLogsProcessed": totalLogsProcessed,
		},
	})
}

// GetParserHandler returns a single parser by its ObjectID or hash.
func (api *Config) GetParserHandler(req *notnet.Request, res *notnet.Response) error {
	ctx := req.HTTPRequest.Context()
	idOrHash := req.Param("id")
	if idOrHash == "" {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "parser id or hash is required",
		})
	}

	var p *database.ParserDoc
	var err error

	if objID, parseErr := bson.ObjectIDFromHex(idOrHash); parseErr == nil {
		p, err = api.Reqs.FindParserByID(ctx, objID)
	} else {
		p, err = api.Reqs.FindParserByHash(ctx, idOrHash)
	}

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.JSON(http.StatusNotFound, map[string]string{
				"error": "parser not found",
			})
		}
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve parser",
		})
	}

	return res.JSON(http.StatusOK, p)
}

// UpdateParserHandler updates the name, format, or status of an existing parser.
func (api *Config) UpdateParserHandler(req *notnet.Request, res *notnet.Response) error {
	ctx := req.HTTPRequest.Context()
	idStr := req.Param("id")
	if idStr == "" {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "parser id is required",
		})
	}

	objID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		// If not ObjectID, check if it's a hash
		doc, findErr := api.Reqs.FindParserByHash(ctx, idStr)
		if findErr != nil {
			return res.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid parser id or hash",
			})
		}
		objID = doc.ID
	}

	type updatePayload struct {
		Name   string `json:"name"`
		Format string `json:"format"`
		Status string `json:"status"`
	}

	var payload updatePayload
	if err := json.NewDecoder(req.HTTPRequest.Body).Decode(&payload); err != nil {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	updated, err := api.Reqs.UpdateParser(ctx, objID, payload.Name, payload.Format, payload.Status)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to update parser",
		})
	}

	return res.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"parser":  updated,
	})
}

// CreateParserHandler allows manual creation of a log parser.
func (api *Config) CreateParserHandler(req *notnet.Request, res *notnet.Response) error {
	ctx := req.HTTPRequest.Context()

	type createPayload struct {
		Name            string                           `json:"name"`
		Format          string                           `json:"format"`
		Pattern         string                           `json:"pattern"`
		Mapping         map[string]string                `json:"mapping"`
		Transformations map[string]parser.Transformation `json:"transformations,omitempty"`
		Status          string                           `json:"status"`
		SampleLog       string                           `json:"sampleLog,omitempty"`
	}

	var payload createPayload
	if err := json.NewDecoder(req.HTTPRequest.Body).Decode(&payload); err != nil {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if payload.Pattern == "" {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "pattern is required",
		})
	}

	if payload.Name == "" {
		if payload.Format != "" {
			payload.Name = payload.Format + "Parser"
		} else {
			payload.Name = "CustomParser"
		}
	}
	if payload.Status == "" {
		payload.Status = "Active"
	}
	if payload.Format == "" {
		payload.Format = "Custom"
	}
	if payload.Mapping == nil {
		payload.Mapping = make(map[string]string)
	}

	// Generate deterministic hash from pattern
	h := sha256.New()
	h.Write([]byte(payload.Pattern))
	hash := hex.EncodeToString(h.Sum(nil))[:16]

	doc := database.ParserDoc{
		Name:            payload.Name,
		Hash:            hash,
		Pattern:         payload.Pattern,
		Mapping:         payload.Mapping,
		Transformations: payload.Transformations,
		Status:          payload.Status,
		Format:          payload.Format,
		SampleLog:       payload.SampleLog,
	}

	saved, err := api.Reqs.SaveParser(ctx, &doc)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to save parser to database",
		})
	}

	return res.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"parser":  saved,
	})
}
