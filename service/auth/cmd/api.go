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
	"ticket-tix/service/auth/internal/handler"
	intRedis "ticket-tix/service/auth/internal/infra/redis"
	"ticket-tix/service/auth/internal/repository"
	"ticket-tix/service/auth/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	httpPort = "50061"

	redisPort = "localhost:6379"

	secretKey = "sudo-secret-key"

	dbHost = "localhost"
	dbPort = 5433
	dbUser = "user"
	dbPass = "password"
	dbName = "ticket_tix_db"
)

func main() {

	dbConn := openDBConnection()
	redisConn := openRedisConnection()

	tokenCache := intRedis.NewRefreshToken(redisConn)

	userRepo := repository.NewUserRepo(dbConn)
	userService := service.NewUserService(userRepo, secretKey, tokenCache)
	userHandler := handler.NewHandler(userService, redisConn)

	r := gin.Default()

	userHandler.RegisterRoutes(r)

	srv := &http.Server{Addr: ":" + httpPort, Handler: r}
	go func() {
		log.Println("starting HTTP server on port " + httpPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}

	redisConn.Close()
	dbConn.Close()
	log.Println("auth service stopped gracefully")
}

func openDBConnection() *sql.DB {
	cfg := db.Config{
		Host: dbHost,
		Port: dbPort,
		User: dbUser,
		Pass: dbPass,
		DB:   dbName,
	}

	pg, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pg.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return pg
}

func openRedisConnection() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisPort,
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
