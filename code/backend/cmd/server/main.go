package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/rabbitmq"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/redis"
)

func main() {
	rdb, err := redis.New(redis.RedisOptions{URL: os.Getenv("REDIS_URL")})
	if err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}

	rmq, err := rabbitmq.New(rabbitmq.RabbitMQOptions{
		URL:          os.Getenv("RABBITMQ_URL"),
		QueueName:    "gologgo_queue",
		RoutingKey:   "gologgo.event",
		ExchangeName: "gologgo_exchange",
	})
	if err != nil {
		log.Fatalf("failed to initialize rabbitmq: %v", err)
	}

	fmt.Println("GoLogGo backend started successfully")
	fmt.Printf("Redis connected: %v\n", os.Getenv("REDIS_URL"))
	fmt.Printf("RabbitMQ connected: %v\n", os.Getenv("RABBITMQ_URL"))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a shutdown signal.
	sig := <-sigCh
	log.Printf("received signal %v, shutting down...", sig)

	// Give ongoing operations some time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	// Currently Close() doesn't take a context, so perform cleanup directly.
	// The context can be passed to HTTP servers/workers later.
	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)

		if err := rmq.Close(); err != nil {
			log.Printf("rabbitmq close error: %v", err)
		}

		if err := rdb.Close(); err != nil {
			log.Printf("redis close error: %v", err)
		}
	}()

	select {
	case <-shutdownDone:
		log.Println("shutdown completed successfully")

	case <-shutdownCtx.Done():
		log.Println("shutdown timed out")
	}

	log.Println("GoLogGo backend stopped")
}
