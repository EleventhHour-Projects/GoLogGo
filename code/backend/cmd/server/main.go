package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/api"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/auth"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/database"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/parsergen"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/rabbitmq"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/redis"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/schedular"
	"github.com/EleventhHour-Projects/GoLogGo/code/backend/internal/worker"

	"github.com/nottechdm/notnet/pkg/notnet"
	"go.mongodb.org/mongo-driver/v2/bson"
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

	// mongodb init
	ctx := context.Background()
	mongod, err := database.New(ctx, database.MongoDBOptions{
		URI:          os.Getenv("MONGODB_URI"),
		DatabaseName: os.Getenv("MONGODB_DATABASE_NAME"),
	})
	if err != nil {
		log.Fatalf("failed to initialize mongodb: %v", err)
	}
	defer mongod.Close(ctx)

	// shared channel for schedular, all workers
	const ChanSize = 1200
	JobChan := make(chan bson.ObjectID, ChanSize)

	// apiCfg init
	apiCfg := api.Config{
		Reqs:    mongod,
		JobChan: JobChan,
	}

	// worker pool init
	numWorkers := 100 // worker-pool size
	workerPool := worker.NewWorker(mongod, rdb, rmq, JobChan, numWorkers)
	go workerPool.InitialiseWorkerPool(ctx) // concurrency core

	// parser generator consumer init
	parserGen := parsergen.New(mongod, rdb, rmq, JobChan)
	go func() {
		if err := parserGen.Start(ctx); err != nil {
			log.Printf("parser generator consumer stopped: %v", err)
		}
	}()

	app := notnet.New(nil)
	app.Use(notnet.CORS(&notnet.CORSConfig{}))
	app.Use(notnet.Logger(), notnet.Recovery())
	app.GET("/health", func(req *notnet.Request, res *notnet.Response) error {
		return res.JSON(200, map[string]string{"status": "ok"})
	})
	app.POST("/user", apiCfg.HandlerUser)
	app.POST("/log", auth.MiddlewareAuth(apiCfg.LogHandler))
	app.GET("/logs", auth.MiddlewareAuth(apiCfg.GetUserLogsHandler))
	app.GET("/parsers", apiCfg.ListParsersHandler)
	app.GET("/parsers/:id", apiCfg.GetParserHandler)
	app.PATCH("/parsers/:id", apiCfg.UpdateParserHandler)
	app.POST("/parsers", apiCfg.CreateParserHandler)

	// schedular init
	sched := schedular.NewSchedular(mongod, JobChan)
	go sched.Run(ctx)

	fmt.Println("GoLogGo backend started successfully")
	fmt.Printf("Redis connected: %v\n", os.Getenv("REDIS_URL"))
	fmt.Printf("RabbitMQ connected: %v\n", os.Getenv("RABBITMQ_URL"))
	log.Println("API server starting on :8080")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Listen(":8080")
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a shutdown signal.
	sig := <-sigCh
	log.Printf("received signal %v, shutting down...", sig)

	if err := app.Shutdown(); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	// Give ongoing operations some time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

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

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("http server error: %v", err)
		}
	default:
	}

	log.Println("GoLogGo backend stopped")
}
