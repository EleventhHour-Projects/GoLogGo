package rabbitmq

import (
	"context"
	"errors"

	"github.com/wagslane/go-rabbitmq"
)

// PublishMessage publishes a message to the RabbitMQ exchange.
//
// It uses the routing key and exchange name specified in the RabbitMQ struct. If the publisher is not initialized, it returns an error.
// The context is used to handle cancellation and timeouts.
func (r *RabbitMQ) PublishMessage(ctx context.Context, message []byte) error {
	if r.Publisher == nil {
		return errors.New("rabbitmq publisher is not initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return r.Publisher.Publish(
		message,
		[]string{r.routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(r.exchangeName),
	)
}

// ConsumeMessages starts consuming messages from RabbitMQ.
//
// The handler should return nil when the message has been processed
// successfully. A non-nil error causes the message to be rejected.
func (r *RabbitMQ) ConsumeMessages(ctx context.Context, handler func(message []byte) error) error {
	if r.Consumer == nil {
		return errors.New("rabbitmq consumer is not initialized")
	}

	if handler == nil {
		return errors.New("message handler cannot be nil")
	}

	return r.Consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		if err := handler(d.Body); err != nil {
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	})
}
