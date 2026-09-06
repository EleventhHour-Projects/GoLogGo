package api

import (
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Config struct {
	Reqs    *database.MongoDB // save whole MongoDB struct
	JobChan chan bson.ObjectID
}
