package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBCollection string

const (
	// CollectionUsers is the name of the users collection in MongoDB.
	CollectionUsers MongoDBCollection = "users"

	// CollectionRequests is the name of the requests collection in MongoDB.
	CollectionRequests MongoDBCollection = "requests"

	// CollectionLogs is the name of the logs collection in MongoDB.
	CollectionLogs MongoDBCollection = "logs"

	// CollectionParsers is the name of the parsers collection in MongoDB.
	CollectionParsers MongoDBCollection = "parsers"
)

type MongoDB struct {
	Client      *mongo.Client
	Database    *mongo.Database
	Collections map[MongoDBCollection]*mongo.Collection
}

type MongoDBOptions struct {
	// URI is the connection string to the MongoDB server.
	URI string

	// DatabaseName is the name of the database to use.
	DatabaseName string
}

// New creates a new MongoDB instance with the given options.
func New(ctx context.Context, config MongoDBOptions) (*MongoDB, error) {
	if config.URI == "" {
		return nil, errors.New("mongodb URI is required")
	}

	if config.DatabaseName == "" {
		return nil, errors.New("mongodb database name is required")
	}

	client, err := mongo.Connect(
		options.Client().ApplyURI(config.URI),
	)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	db := client.Database(config.DatabaseName)

	return &MongoDB{
		Client:   client,
		Database: db,
		Collections: map[MongoDBCollection]*mongo.Collection{
			CollectionUsers:    db.Collection(string(CollectionUsers)),
			CollectionRequests: db.Collection(string(CollectionRequests)),
			CollectionLogs:     db.Collection(string(CollectionLogs)),
			CollectionParsers:  db.Collection(string(CollectionParsers)),
		},
	}, nil
}

// Close closes the MongoDB connection.
func (m *MongoDB) Close(ctx context.Context) error {
	if m.Client == nil {
		return nil
	}

	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) GetPendingRequestIDs(ctx context.Context) ([]bson.ObjectID, error) {
	collection := m.Collections[CollectionRequests]

	filter := bson.M{
		"status":   StatusPending,
		"attempts": bson.M{"$lt": 3},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []bson.ObjectID

	for cursor.Next(ctx) {
		var request struct {
			ID bson.ObjectID `bson:"_id"`
		}

		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}

		ids = append(ids, request.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
