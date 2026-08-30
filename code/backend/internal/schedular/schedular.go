package schedular

import (
	"context"
	"log"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Schedular struct {
	ReqChan chan bson.ObjectID // channel to push request IDs in
	Reqs    *database.MongoDB  // requests collection
}

func NewSchedular(reqs *database.MongoDB, reqChan chan bson.ObjectID) *Schedular {
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
	ids, err := s.Reqs.GetPendingRequestIDs(ctx)
	if err != nil {
		log.Println("failed to get pending requests:", err)
		return err
	}

	for _, id := range ids {
		s.ReqChan <- id
	}
	return nil
}
