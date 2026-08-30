package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/auth"
	"github.com/nottechdm/notnet/pkg/notnet"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// HandlerUser extracts Name, Email from body, generates JWTTokenString and saves the user to MongoDB Database
func (api *Config) HandlerUser(req *notnet.Request, res *notnet.Response) error {
	type parameters struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	params := parameters{}
	err := json.NewDecoder(req.HTTPRequest.Body).Decode(&params)
	if err != nil {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid Request Body",
		})
	}

	if params.Name == "" || params.Email == "" {
		return res.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name and email are required",
		})
	}

	ctx := req.HTTPRequest.Context()
	user, err := api.Reqs.FindUserByEmail(ctx, params.Email)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// create user
			user, err = api.Reqs.CreateUser(ctx, params.Name, params.Email)
			if err != nil {
				return res.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error Creating User",
				})
			}
		} else {
			return res.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error Finding User",
			})
		}
	}

	// generate token
	tokenString, err := auth.GenerateToken(user.ID)
	if err != nil {
		return res.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error Generating Token",
		})
	}

	// return email and tokenString
	type response struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}

	return res.JSON(http.StatusCreated, response{
		Email: user.Email,
		Token: tokenString,
	})
}
