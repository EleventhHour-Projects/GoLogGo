package schedular

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Schedular struct {
	ReqChan chan bson.ObjectID // channel to push request IDs in
	Reqs    *mongo.Collection  // requests collection
}

func NewSchedular(reqs *mongo.Collection, reqChan chan bson.ObjectID) *Schedular {
	return &Schedular{
		ReqChan: reqChan,
		Reqs:    reqs,
	}
}

// Runs every 5sec, queries database for requests with state = "PENDING" and attempts < 3, pushes ID to ReqChan
func (s *Schedular) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.findAndPushPendingReqs(ctx)
			if err != nil {
				log.Println("Error pushing request IDs to Request Channel: ", err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

// queries db, finds reqs with state = "PENDING" and attempts < 3, pushes them to ReqChan
func (s *Schedular) findAndPushPendingReqs(ctx context.Context) error {
	filter := bson.M{
		"status":   "PENDING",
		"attempts": bson.M{"$lt": 3},
	}

	cursor, err := s.Reqs.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var reqIDs []bson.ObjectID
	cursor.All(ctx, &reqIDs)
	for _, reqID := range reqIDs {
		s.ReqChan <- reqID
	}
	return nil
}
