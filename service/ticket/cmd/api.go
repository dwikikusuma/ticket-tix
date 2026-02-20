package main

import (
	"database/sql"
	"log"
	"ticket-tix/common/pkg/db"
	"ticket-tix/service/ticket/internal/handler"

	"github.com/gin-gonic/gin"
)

const (
	dbHost = "localhost"
	dbPort = 5432
	dbUser = "user"
	dbPass = "password"
	dbName = "ticket_tix_db"
	port   = "50061"
)

func main() {
	log.Println("ticket-tix start")

	ticketDB := openDBConnection()
	defer ticketDB.Close()

	r := gin.Default()

	ticketHandler := handler.NewTicketHandler()
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
		log.Println("failed to open db")
		log.Fatalf(err.Error())
	}
	log.Println("db opened")

	return postgresDB
}
