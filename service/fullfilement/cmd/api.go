package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"ticket-tix/common/pkg/events"
	"ticket-tix/service/fullfilement/internal/handler"
	"ticket-tix/service/fullfilement/internal/repo"
	"ticket-tix/service/fullfilement/internal/servcie"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	client, err := mongo.Connect(startupCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		logger.Error("failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}

	if err := client.Ping(startupCtx, nil); err != nil {
		logger.Error("failed to ping MongoDB", "err", err)
		os.Exit(1)
	}
	logger.Info("connected to MongoDB")

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error("failed to disconnect MongoDB", "err", err)
		}
		logger.Info("disconnected from MongoDB")
	}()

	db := client.Database("ticket-tix")
	collection := db.Collection("fulfillments")

	nosqlDB := repo.NewRepository(collection)
	fulfillmentService := servcie.NewService(nosqlDB)
	h := handler.NewHandler(logger, fulfillmentService)

	router := events.NewRouter()
	router.Handle(
		"booking.created",
		h.HandleOrderCreated,
		events.WithLogging(logger),
		events.WithRetry(3, 500*time.Millisecond),
	)

	consumerCfg := events.DefaultConsumerConfig([]string{"localhost:9092"}, "fulfillment")
	consumer, err := events.NewConsumer(consumerCfg, router, logger)
	if err != nil {
		logger.Error("failed to create consumer", "err", err)
		os.Exit(1)
	}
	defer consumer.Close()

	logger.Info("fulfillment service started")
	if err := consumer.Start(appCtx); err != nil {
		logger.Error("consumer exited with error", "err", err)
		os.Exit(1)
	}
	logger.Info("fulfillment service stopped gracefully")
}
