package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBCollection string

const (
	// CollectionUsers is the name of the users collection in MongoDB.
	CollectionUsers MongoDBCollection = "users"

	// CollectionRequests is the name of the requests collection in MongoDB.
	CollectionRequests MongoDBCollection = "requests"
)

type MongoDB struct {
	Client      *mongo.Client
	Database    *mongo.Database
	Collections map[MongoDBCollection]*Collection
}

type Collection struct {
	Collection *mongo.Collection
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
		Collections: map[MongoDBCollection]*Collection{
			CollectionUsers: {
				Collection: db.Collection(string(CollectionUsers)),
			},
			CollectionRequests: {
				Collection: db.Collection(string(CollectionRequests)),
			},
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
