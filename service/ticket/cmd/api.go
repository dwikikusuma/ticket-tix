package main

import (
	"database/sql"
	"log"
	"ticket-tix/common/pkg/db"
	"ticket-tix/common/pkg/storage"
	"ticket-tix/service/ticket/internal/handler"
	"ticket-tix/service/ticket/internal/repository"
	"ticket-tix/service/ticket/internal/service"

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
	port = "50061"

	// minio
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin123"
	minioBucket    = "ticket-bucket	"
	minioUseSSL    = false
)

func main() {
	log.Println("ticket-tix start")

	ticketDB := openDBConnection()
	defer ticketDB.Close()

	minioStorage := openStorageConnection()

	ticketRepo := repository.NewTicketRepo(ticketDB)
	ticketService := service.NewTicketService(ticketDB, minioStorage, ticketRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)

	r := gin.Default()
	ticketHandler.RegisterRoutes(r)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func openDBConnection() *sql.DB {
	log.Println("opening db connection")
	dbConfig := db.Config{
		Host:            dbHost,
		Port:            dbPort,
		User:            dbUser,
		Pass:            dbPass,
		DB:              dbName,
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30,
	}

	postgresDB, err := db.Open(dbConfig)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	log.Println("db opened")
	return postgresDB
}

func openStorageConnection() *storage.Storage {
	log.Println("opening minio connection")
	storageCfg := storage.StorageConfig{
		Endpoint:        minioEndpoint,
		AccessKey:       minioAccessKey,
		SecretAccessKey: minioSecretKey,
		UseSSL:          minioUseSSL,
		BucketName:      minioBucket,
	}

	s, err := storage.NewStorage(storageCfg)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}
	log.Println("minio connected")
	return s
}
