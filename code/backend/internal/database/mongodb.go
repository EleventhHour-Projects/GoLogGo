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

// Req represents a request in the MongoDB database.
type Req struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Payload   []byte        `bson:"payload" json:"payload"`
	Status    string        `bson:"status" json:"status"`
	Attempts  int           `bson:"attempts" json:"attempts"`
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

// CreateReq creates a new request in the MongoDB database.
func (m *MongoDB) CreateReq(ctx context.Context, payload []byte) (*Req, error) {
	now := time.Now()
	req := &Req{
		ID:        bson.NewObjectID(),
		Payload:   payload,
		Status:    "pending",
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
