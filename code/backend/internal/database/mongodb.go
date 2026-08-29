package database

import (
	"context"
	"time"

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
