package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	var wg sync.WaitGroup
	httpServer := spinUpHttpServer(r, &wg)

	log.Println("all services started")
	sigChannel := make(chan os.Signal, 1)

	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	<-sigChannel

	stopHttpServer(httpServer)
	wg.Wait()
	log.Println("all services stopped gracefully")
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

func spinUpHttpServer(r *gin.Engine, wg *sync.WaitGroup) *http.Server {
	svr := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("starting HTTP server on port " + httpPort)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	return svr
}

func stopHttpServer(svr *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		log.Println("HTTP server forced to shutdown: %v", err)

	} else {
		log.Println("HTTP server stopped gracefully")
	}
}
