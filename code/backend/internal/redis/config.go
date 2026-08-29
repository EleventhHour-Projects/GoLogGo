package redis

import (
	"context"
	"errors"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *goredis.Client
}

type RedisOptions struct {
	// URL is the connection string to the Redis server.
	URL string
	// DB is the Redis database number to use.
	DB int
}

const (
	// ErrMissingURL is returned when the URL is not provided in the options.
	ErrMissingURL = "missing URL in Redis options"
)

// New creates a new Redis instance with the given options.
func New(options RedisOptions) (*Redis, error) {
	if options.URL == "" {
		return nil, errors.New(ErrMissingURL)
	}

	clientOptions, err := goredis.ParseURL(options.URL)
	if err != nil {
		return nil, err
	}

	if options.DB > 0 {
		clientOptions.DB = options.DB
	}

	client := goredis.NewClient(clientOptions)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Redis{Client: client}, nil
}

// Close closes the Redis connection.
func (r *Redis) Close() error {
	if r == nil || r.Client == nil {
		return nil
	}

	return r.Client.Close()
}
