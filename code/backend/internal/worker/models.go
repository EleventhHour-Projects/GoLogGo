package worker

import (
	"context"
)

// MessagePublisher abstracts publishing messages to a message broker (e.g. RabbitMQ).
type MessagePublisher interface {
	PublishMessage(ctx context.Context, message []byte) error
}
