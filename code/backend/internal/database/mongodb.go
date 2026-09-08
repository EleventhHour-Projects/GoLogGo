package database

import (
	"context"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parser"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a user in the MongoDB database.
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Name      string        `bson:"name" json:"name"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// Req represents a request in the MongoDB database.
type Req struct {
	ID        bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID    *bson.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	Payload   []byte         `bson:"payload" json:"payload"`
	Status    ReqestStatus   `bson:"status" json:"status"`
	Attempts  int            `bson:"attempts" json:"attempts"`
	CreatedAt time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time      `bson:"updatedAt" json:"updatedAt"`
}

// NormalizedLog is an alias to parser.NormalizedLog.
type NormalizedLog = parser.NormalizedLog

// Log represents a log entry in the MongoDB database.
type Log struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id"`

	// UserID is a reference to the user who created the log entry. It can be nil if the log entry was created by the system.
	UserID *bson.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	// IsDirect indicates whether the log entry was created directly by the user or directly through current machine.
	IsDirect bool `bson:"isDirect" json:"isDirect"`

	ReqID            bson.ObjectID `bson:"reqId" json:"reqId"`
	RequestCreatedAt time.Time     `bson:"requestCreatedAt" json:"requestCreatedAt"`

	NormalizedLog NormalizedLog `bson:"normalizedLog" json:"normalizedLog"`
	RawLog        []byte        `bson:"rawLog" json:"rawLog"`
	Hash          string        `bson:"hash" json:"hash"`
	CreatedAt     time.Time     `bson:"createdAt" json:"createdAt"`
}

// ReqestStatus represents the status of a request in the MongoDB database.
type ReqestStatus string

// Define constants for request statuses
const (
	StatusPending       ReqestStatus = "pending"
	StatusProcessing    ReqestStatus = "processing"
	StatusWaitingParser ReqestStatus = "waiting_parser"
	StatusCompleted     ReqestStatus = "completed"
	StatusFailed        ReqestStatus = "failed"
)

// CreateUser creates a new user in the MongoDB database.
func (m *MongoDB) CreateUser(ctx context.Context, name, email string) (*User, error) {
	now := time.Now()
	user := &User{
		ID:        bson.NewObjectID(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	collection := m.Collections[CollectionUsers]

	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateReq creates a new request in the MongoDB database.
func (m *MongoDB) CreateReq(ctx context.Context, payload []byte, userID *bson.ObjectID) (*Req, error) {
	now := time.Now()
	req := &Req{
		ID:        bson.NewObjectID(),
		UserID:    userID,
		Payload:   payload,
		Status:    StatusPending,
		Attempts:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	collection := m.Collections[CollectionRequests]

	_, err := collection.InsertOne(ctx, req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// FindUserByEmail finds if a user already exists in the database with given Email
func (m *MongoDB) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	collection := m.Collections[CollectionUsers]

	var user User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *MongoDB) FindReqByID(ctx context.Context, id bson.ObjectID) (*Req, error) {
	collection := m.Collections[CollectionRequests]

	var req Req
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

// UpdateReqStatus updates the status of a request in the MongoDB database.
func (m *MongoDB) UpdateReqStatus(ctx context.Context, id bson.ObjectID, status ReqestStatus) error {
	collection := m.Collections[CollectionRequests]

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status, "updatedAt": time.Now()}})
	if err != nil {
		return err
	}

	return nil
}

// IncrementReqAttempts increments the attempts count of a request in the MongoDB database.
func (m *MongoDB) IncrementReqAttempts(ctx context.Context, id bson.ObjectID) error {
	collection := m.Collections[CollectionRequests]

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"attempts": 1}, "$set": bson.M{"updatedAt": time.Now()}})
	if err != nil {
		return err
	}

	return nil
}

// DeleteReq deletes a request from the MongoDB database.
func (m *MongoDB) DeleteReq(ctx context.Context, id bson.ObjectID) error {
	collection := m.Collections[CollectionRequests]

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	return nil
}

// InsertLog inserts a new log entry into the MongoDB database.
func (m *MongoDB) InsertLog(ctx context.Context, req Req, nlog NormalizedLog, isdirect bool, userID *bson.ObjectID, hash string) error {
	collection := m.Collections[CollectionLogs]

	log := &Log{
		ID:               bson.NewObjectID(),
		UserID:           userID,
		IsDirect:         isdirect,
		ReqID:            req.ID,
		RequestCreatedAt: req.CreatedAt,
		NormalizedLog:    nlog,
		RawLog:           req.Payload,
		Hash:             hash,
		CreatedAt:        time.Now(),
	}

	_, err := collection.InsertOne(ctx, log)
	if err != nil {
		return err
	}

	return nil
}

// ParserDoc represents a parser document stored in MongoDB.
type ParserDoc struct {
	ID              bson.ObjectID                    `bson:"_id,omitempty" json:"id"`
	Name            string                           `bson:"name" json:"name"`
	Hash            string                           `bson:"hash" json:"hash"`
	Pattern         string                           `bson:"pattern" json:"pattern"`
	Mapping         map[string]string                `bson:"mapping" json:"mapping"`
	Transformations map[string]parser.Transformation `bson:"transformations,omitempty" json:"transformations,omitempty"`
	Status          string                           `bson:"status" json:"status"`
	Format          string                           `bson:"format" json:"format"`
	Vendor          string                           `bson:"vendor,omitempty" json:"vendor,omitempty"`
	Product         string                           `bson:"product,omitempty" json:"product,omitempty"`
	Template        string                           `bson:"template,omitempty" json:"template,omitempty"`
	Delimiter       string                           `bson:"delimiter,omitempty" json:"delimiter,omitempty"`
	ExtractedFields []string                         `bson:"extractedFields,omitempty" json:"extractedFields,omitempty"`
	FieldSignatures []map[string]string              `bson:"fieldSignatures,omitempty" json:"fieldSignatures,omitempty"`
	SampleLog       string                           `bson:"sampleLog,omitempty" json:"sampleLog,omitempty"`
	LogsProcessed   int64                            `bson:"logsProcessed" json:"logsProcessed"`
	LastUsed        time.Time                        `bson:"lastUsed" json:"lastUsed"`
	CreatedAt       time.Time                        `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time                        `bson:"updatedAt" json:"updatedAt"`
}

// SaveParser inserts or updates a parser in MongoDB by Hash.
func (m *MongoDB) SaveParser(ctx context.Context, p *ParserDoc) (*ParserDoc, error) {
	collection := m.Collections[CollectionParsers]
	now := time.Now()

	if p.Name == "" {
		if p.Product != "" {
			p.Name = p.Product + "Parser"
		} else if p.Vendor != "" {
			p.Name = p.Vendor + "Parser"
		} else if p.Format != "" && p.Format != "unknown" {
			p.Name = p.Format + "Parser"
		} else if len(p.Hash) >= 8 {
			p.Name = "Parser-" + p.Hash[:8]
		} else {
			p.Name = "CustomParser"
		}
	}
	if p.Status == "" {
		p.Status = "Active"
	}
	if p.Format == "" {
		p.Format = "Custom"
	}

	var existing ParserDoc
	err := collection.FindOne(ctx, bson.M{"hash": p.Hash}).Decode(&existing)
	if err == nil {
		// Update existing
		updateFields := bson.M{
			"pattern":         p.Pattern,
			"mapping":         p.Mapping,
			"transformations": p.Transformations,
			"format":          p.Format,
			"updatedAt":       now,
		}
		if p.Vendor != "" {
			updateFields["vendor"] = p.Vendor
		}
		if p.Product != "" {
			updateFields["product"] = p.Product
		}
		if p.Template != "" {
			updateFields["template"] = p.Template
		}
		if p.Delimiter != "" {
			updateFields["delimiter"] = p.Delimiter
		}
		if len(p.ExtractedFields) > 0 {
			updateFields["extractedFields"] = p.ExtractedFields
		}
		if len(p.FieldSignatures) > 0 {
			updateFields["fieldSignatures"] = p.FieldSignatures
		}
		if p.SampleLog != "" {
			updateFields["sampleLog"] = p.SampleLog
		}
		// If existing parser had a custom name, preserve it unless explicitly overwritten
		if existing.Name != "" && (p.Name == "" || p.Name == "CustomParser" || (len(p.Hash) >= 8 && p.Name == "Parser-"+p.Hash[:8])) {
			p.Name = existing.Name
		} else if p.Name != "" {
			updateFields["name"] = p.Name
		}
		if p.Status != "" {
			updateFields["status"] = p.Status
		}

		_, err = collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": updateFields})
		if err != nil {
			return nil, err
		}
		existing.Pattern = p.Pattern
		existing.Mapping = p.Mapping
		existing.Transformations = p.Transformations
		existing.Format = p.Format
		existing.Vendor = p.Vendor
		existing.Product = p.Product
		existing.Template = p.Template
		existing.Delimiter = p.Delimiter
		existing.ExtractedFields = p.ExtractedFields
		existing.FieldSignatures = p.FieldSignatures
		existing.UpdatedAt = now
		if p.SampleLog != "" {
			existing.SampleLog = p.SampleLog
		}
		return &existing, nil
	}

	// Insert new
	p.ID = bson.NewObjectID()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.LastUsed = now

	_, err = collection.InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// FindParserByHash finds a parser by its hash.
func (m *MongoDB) FindParserByHash(ctx context.Context, hash string) (*ParserDoc, error) {
	collection := m.Collections[CollectionParsers]
	var p ParserDoc
	err := collection.FindOne(ctx, bson.M{"hash": hash}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindParserByID finds a parser by its ObjectID.
func (m *MongoDB) FindParserByID(ctx context.Context, id bson.ObjectID) (*ParserDoc, error) {
	collection := m.Collections[CollectionParsers]
	var p ParserDoc
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// FindAllParsers retrieves all parsers.
func (m *MongoDB) FindAllParsers(ctx context.Context) ([]ParserDoc, error) {
	collection := m.Collections[CollectionParsers]
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var parsers []ParserDoc
	for cursor.Next(ctx) {
		var p ParserDoc
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		parsers = append(parsers, p)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if parsers == nil {
		parsers = []ParserDoc{}
	}

	return parsers, nil
}

// UpdateParser updates parser name, format, or status by ID.
func (m *MongoDB) UpdateParser(ctx context.Context, id bson.ObjectID, name, format, status string) (*ParserDoc, error) {
	collection := m.Collections[CollectionParsers]
	update := bson.M{"updatedAt": time.Now()}

	if name != "" {
		update["name"] = name
	}
	if format != "" {
		update["format"] = format
	}
	if status != "" {
		update["status"] = status
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}

	return m.FindParserByID(ctx, id)
}

// IncrementParserLogsProcessed increments logsProcessed count and sets lastUsed timestamp.
func (m *MongoDB) IncrementParserLogsProcessed(ctx context.Context, hash string) error {
	collection := m.Collections[CollectionParsers]
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"hash": hash},
		bson.M{
			"$inc": bson.M{"logsProcessed": 1},
			"$set": bson.M{"lastUsed": time.Now()},
		},
	)
	return err
}

