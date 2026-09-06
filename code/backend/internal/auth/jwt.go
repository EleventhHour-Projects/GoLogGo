package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GenerateToken(userID bson.ObjectID) (string, error) {
	secret := os.Getenv("JWT_KEY")
	if secret == "" {
		return "", fmt.Errorf("JWT_KEY not set")
	}

	jwtKey := []byte(secret)

	// expiration time ~2 months
	expirationTime := time.Now().Add(1440 * time.Hour)

	claims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-auth-internal",
		},
	}

	// Create token with claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates the token string
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// Parsing logic ensures the token is signed with the expected method (HMAC)
	secret := os.Getenv("JWT_KEY")
	if secret == "" {
		return nil, fmt.Errorf("JWT_KEY not set")
	}

	jwtKey := []byte(secret)

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Return validated claims or an error
	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Issuer != "go-auth-internal" {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}
