package api

import (
	"net/http"
	"strconv"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/auth"
	"github.com/nottechdm/notnet/pkg/notnet"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetUserLogsHandler returns paginated requests + stats for the authenticated user.
// GET /logs?page=1&limit=25&status=all
func (api *Config) GetUserLogsHandler(req *notnet.Request, res *notnet.Response, claims *auth.CustomClaims) error {
	ctx := req.HTTPRequest.Context()

	// Parse query params
	query := req.HTTPRequest.URL.Query()
	page, _ := strconv.ParseInt(query.Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit < 1 || limit > 100 {
		limit = 25
	}
	status := query.Get("status") // e.g. "completed", "pending", "" / "all"

	skip := (page - 1) * limit

	// Fetch stats
	stats, err := api.Reqs.GetUserStats(ctx, claims.UserID)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch user stats",
		})
	}

	// Fetch paginated requests
	reqs, total, err := api.Reqs.GetUserRequests(ctx, claims.UserID, limit, skip, status)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch requests",
		})
	}

	// For completed requests, fetch their parsed logs
	type ReqRow struct {
		ID        string      `json:"id"`
		Payload   string      `json:"payload"`
		Status    string      `json:"status"`
		Attempts  int         `json:"attempts"`
		CreatedAt interface{} `json:"createdAt"`
		UpdatedAt interface{} `json:"updatedAt"`
		Logs      interface{} `json:"logs,omitempty"`
	}

	rows := make([]ReqRow, 0, len(reqs))
	for _, r := range reqs {
		row := ReqRow{
			ID:        r.ID.Hex(),
			Payload:   string(r.Payload),
			Status:    string(r.Status),
			Attempts:  r.Attempts,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
		if r.Status == "completed" {
			logs, err := api.Reqs.GetLogsByReqID(ctx, r.ID)
			if err == nil {
				row.Logs = logs
			}
		}
		rows = append(rows, row)
	}

	return res.JSON(http.StatusOK, map[string]interface{}{
		"stats":      stats,
		"requests":   rows,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + limit - 1) / limit,
	})
}

// GetUserLogHandler returns the original request and its parsed log entries.
// GET /logs/:id
func (api *Config) GetUserLogHandler(req *notnet.Request, res *notnet.Response, claims *auth.CustomClaims) error {
	requestID, err := bson.ObjectIDFromHex(req.Param("id"))
	if err != nil {
		return res.JSON(http.StatusBadRequest, map[string]string{"error": "invalid log id"})
	}

	ctx := req.HTTPRequest.Context()
	storedReq, err := api.Reqs.GetUserRequestByID(ctx, claims.UserID, requestID)
	if err != nil {
		return res.JSON(http.StatusNotFound, map[string]string{"error": "log not found"})
	}

	logs, err := api.Reqs.GetLogsByReqID(ctx, requestID)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch log details"})
	}

	return res.JSON(http.StatusOK, map[string]interface{}{
		"request": map[string]interface{}{
			"id":        storedReq.ID.Hex(),
			"payload":   string(storedReq.Payload),
			"status":    storedReq.Status,
			"attempts":  storedReq.Attempts,
			"createdAt": storedReq.CreatedAt,
			"updatedAt": storedReq.UpdatedAt,
		},
		"logs": logs,
	})
}
