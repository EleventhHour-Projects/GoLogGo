package rabbitmq

import (
	"errors"

	"github.com/wagslane/go-rabbitmq"
)

type RabbitMQ struct {
	Conn      *rabbitmq.Conn
	Consumer  *rabbitmq.Consumer
	Publisher *rabbitmq.Publisher
	// Settings for the consumer and publisher.
	routingKey   string
	exchangeName string
	queueName    string
}

type RabbitMQOptions struct {
	// URL is the connection string to the RabbitMQ server.
	URL string
	// Below are the options for the consumer.
	QueueName  string
	RoutingKey string
	// Below are the options for the both consumer and publisher.
	ExchangeName string
}

const (
	// ErrMissingURL is returned when the URL is not provided in the options.
	ErrMissingURL = "missing URL in RabbitMQ options"
	// ErrMissingQueueName is returned when the QueueName is not provided in the options.
	ErrMissingQueueName = "missing QueueName in RabbitMQ options"
	// ErrMissingConsumerRoutingKey is returned when the ConsumerRoutingKey is not provided in the options.
	ErrMissingConsumerRoutingKey = "missing ConsumerRoutingKey in RabbitMQ options"
	// ErrMissingExchangeName is returned when the ExchangeName is not provided in the options.
	ErrMissingExchangeName = "missing ExchangeName in RabbitMQ options"
)

// New creates a new RabbitMQ instance with the given options.
//
// It establishes a connection to the RabbitMQ server, creates a consumer and a publisher, and returns an error if any of these steps fail.
func New(options RabbitMQOptions) (*RabbitMQ, error) {
	if options.URL == "" {
		return nil, errors.New(ErrMissingURL)
	}
	if options.QueueName == "" {
		return nil, errors.New(ErrMissingQueueName)
	}
	if options.RoutingKey == "" {
		return nil, errors.New(ErrMissingConsumerRoutingKey)
	}
	if options.ExchangeName == "" {
		return nil, errors.New(ErrMissingExchangeName)
	}

	r := &RabbitMQ{
		exchangeName: options.ExchangeName,
		queueName:    options.QueueName,
		routingKey:   options.RoutingKey,
	}

	conn, err := rabbitmq.NewConn(
		options.URL,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, err
	}
	r.Conn = conn
	consumer, err := rabbitmq.NewConsumer(
		conn,
		options.QueueName,
		rabbitmq.WithConsumerOptionsRoutingKey(options.RoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(options.ExchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		r.Close()
		return nil, err
	}
	r.Consumer = consumer
	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(options.ExchangeName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		r.Close()
		return nil, err
	}
	r.Publisher = publisher
	return r, nil
}

// Close closes the RabbitMQ connection, consumer, and publisher.
func (r *RabbitMQ) Close() error {
	if r.Consumer != nil {
		r.Consumer.Close()
	}

	if r.Publisher != nil {
		r.Publisher.Close()
	}

	if r.Conn != nil {
		r.Conn.Close()
	}
	return nil
}
