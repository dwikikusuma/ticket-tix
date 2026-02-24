package main

import (
	"context"
	"database/sql"
	"log"
	"ticket-tix/common/pkg/db"
	"ticket-tix/service/bookings/internal/handler"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// db
	dbHost = "localhost"
	dbPort = 5433
	dbUser = "user"
	dbPass = "password"
	dbName = "ticket_tix_db"

	// server
	port = "50062"
)

func main() {
	bookDB := openDBConnection()
	defer bookDB.Close()

	httpHandler := handler.NewHandler()

	r := gin.Default()
	httpHandler.RegisterRoutes(r)

	if serveErr := r.Run(":" + port); serveErr != nil {
		log.Fatalf("Failed to start server: %v", serveErr)
	}
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
