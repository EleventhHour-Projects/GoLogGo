package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomClaims struct {
	UserID bson.ObjectID `bson:"_id"`
	jwt.RegisteredClaims
}