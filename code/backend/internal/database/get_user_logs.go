package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ReqWithLogs is a joined view of a request and its parsed logs.
type ReqWithLogs struct {
	Req  Req   `json:"req"`
	Logs []Log `json:"logs"`
}

// GetUserRequests returns paginated requests belonging to a user, newest first.
// Pass a zero ObjectID to get requests across all users (admin use).
func (m *MongoDB) GetUserRequests(
	ctx context.Context,
	userID bson.ObjectID,
	limit, skip int64,
	status string,
) ([]Req, int64, error) {
	col := m.Collections[CollectionRequests]

	filter := bson.M{"userId": userID}
	if status != "" && status != "all" {
		filter["status"] = ReqestStatus(status)
	}

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reqs []Req
	for cursor.Next(ctx) {
		var r Req
		if err := cursor.Decode(&r); err != nil {
			return nil, 0, err
		}
		reqs = append(reqs, r)
	}
	if reqs == nil {
		reqs = []Req{}
	}
	return reqs, total, cursor.Err()
}

// GetLogsByReqID returns all parsed log entries for a given request ID.
func (m *MongoDB) GetLogsByReqID(ctx context.Context, reqID bson.ObjectID) ([]Log, error) {
	col := m.Collections[CollectionLogs]

	cursor, err := col.Find(ctx, bson.M{"reqId": reqID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []Log
	for cursor.Next(ctx) {
		var l Log
		if err := cursor.Decode(&l); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []Log{}
	}
	return logs, cursor.Err()
}

// GetUserStats returns counts of requests per status for a given user.
type UserStats struct {
	Total          int64     `json:"total"`
	Pending        int64     `json:"pending"`
	Processing     int64     `json:"processing"`
	WaitingParser  int64     `json:"waitingParser"`
	Completed      int64     `json:"completed"`
	Failed         int64     `json:"failed"`
	LastActivityAt time.Time `json:"lastActivityAt"`
}

func (m *MongoDB) GetUserStats(ctx context.Context, userID bson.ObjectID) (*UserStats, error) {
	col := m.Collections[CollectionRequests]

	statuses := []ReqestStatus{
		StatusPending, StatusProcessing, StatusWaitingParser, StatusCompleted, StatusFailed,
	}

	stats := &UserStats{}
	for _, s := range statuses {
		n, err := col.CountDocuments(ctx, bson.M{"userId": userID, "status": s})
		if err != nil {
			return nil, err
		}
		switch s {
		case StatusPending:
			stats.Pending = n
		case StatusProcessing:
			stats.Processing = n
		case StatusWaitingParser:
			stats.WaitingParser = n
		case StatusCompleted:
			stats.Completed = n
		case StatusFailed:
			stats.Failed = n
		}
		stats.Total += n
	}

	// Grab the most recent request timestamp
	var latest struct {
		CreatedAt time.Time `bson:"createdAt"`
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if err := col.FindOne(ctx, bson.M{"userId": userID}, opts).Decode(&latest); err == nil {
		stats.LastActivityAt = latest.CreatedAt
	}

	return stats, nil
}
