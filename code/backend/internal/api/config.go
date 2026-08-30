package api

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
)

type Config struct {
	Reqs *database.MongoDB  // save whole MongoDB struct
	JobChan chan bson.ObjectID
}