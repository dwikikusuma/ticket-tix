package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"ticket-tix/common/pkg/db"
	"ticket-tix/common/pkg/events"
	"ticket-tix/common/pkg/lock"
	"ticket-tix/service/bookings/internal/handler"
	"ticket-tix/service/bookings/internal/repository"
	"ticket-tix/service/bookings/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ticketRPC "ticket-tix/common/gen/ticket/v1"
)

const (
	// db
	dbHost = "localhost"
	dbPort = 5433
	dbUser = "user"
	dbPass = "password"
	dbName = "ticket_tix_db"

	// server
	port = "50063"

	// rpc
	ticketRPCAddr = "localhost:40061"

	// kafka
	kafkaAddr = "localhost:9092"
)

func main() {
	bookDB := openDBConnection()

	ticketConn, err := grpc.NewClient(ticketRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to ticket service: %v", err)
	}

	redisClient := openRedisConnection("localhost:6379")
	redLock := lock.NewRedLock(redisClient, redisClient, redisClient)

	ticketClient := ticketRPC.NewTicketServiceClient(ticketConn)
	repo := repository.NewBookingRepo(bookDB)

	producerConfig := events.GetDefaultConfig([]string{kafkaAddr})
	producer, producerErr := events.NewProducer(producerConfig)

	if producerErr != nil {
		log.Fatalf("Failed to create producer: %v", producerErr)
	}

	ticketService := service.NewBookingService(repo, ticketClient, redLock, producer)
	httpHandler := handler.NewHandler(ticketService, redisClient)

	r := gin.Default()
	httpHandler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	log.Println("shut down signal received....")
	log.Println("shutting down HTTP server...")

	shutDownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		log.Printf("HTTP server graceful shutdown failed: %v", err)
		if closeErr := srv.Close(); closeErr != nil {
			log.Printf("HTTP server force close failed: %v", closeErr)
		}
	}

	producer.Close()

	if err := ticketConn.Close(); err != nil {
		log.Printf("Ticket gRPC connection close failed: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Redis connection close failed: %v", err)
	}

	if err := bookDB.Close(); err != nil {
		log.Printf("DB connection close failed: %v", err)
	}

	log.Println("booking service stopped..")

}

func openDBConnection() *sql.DB {
	cfg := db.Config{
		Host: dbHost,
		Port: dbPort,
		User: dbUser,
		Pass: dbPass,
		DB:   dbName,

		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30,
	}

	pg, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if pingErr := pg.PingContext(ctx); pingErr != nil {
		log.Fatalf("Failed to ping database: %v", pingErr)
	}
	return pg
}

func openRedisConnection(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}
