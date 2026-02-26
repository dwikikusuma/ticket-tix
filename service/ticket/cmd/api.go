package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"ticket-tix/common/pkg/db"
	"ticket-tix/common/pkg/storage"
	"ticket-tix/service/ticket/internal/handler"
	"ticket-tix/service/ticket/internal/repository"
	"ticket-tix/service/ticket/internal/service"
	"time"

	ticketRPC "ticket-tix/common/gen/ticket/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	// db
	dbHost = "localhost"
	dbPort = 5433
	dbUser = "user"
	dbPass = "password"
	dbName = "ticket_tix_db"

	// server
	httpPort = "50061"

	// minio
	minioEndpoint  = "localhost:9000"
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin123"
	minioBucket    = "ticket-bucket"
	minioUseSSL    = false

	// rpc
	rpcAddr = "40061"
)

func main() {
	log.Println("ticket-tix start")

	ticketDB := openDBConnection()
	defer ticketDB.Close()

	minioStorage := openStorageConnection()

	ticketRepo := repository.NewTicketRepo(ticketDB)
	ticketService := service.NewTicketService(ticketDB, minioStorage, ticketRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)

	grpcServer := grpc.NewServer()
	rpcHandler := handler.NewRPCHandler(ticketService)
	ticketRPC.RegisterTicketServiceServer(grpcServer, rpcHandler)
	reflection.Register(grpcServer)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	ticketHandler.RegisterRoutes(r)

	// centralized signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	httpServer := spinUpHTTPServer(ctx, r, &wg)
	spinUpGRPCServer(grpcServer, &wg)

	log.Println("all services started")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	<-sigChannel

	log.Println("shutdown signal received, stopping servers...")
	cancel()

	stopHTTPServer(httpServer)
	stopGRPCServer(grpcServer)

	wg.Wait()
	log.Println("all servers stopped, bye!")
}

func spinUpHTTPServer(ctx context.Context, r *gin.Engine, wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("starting HTTP server on port %s", httpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to run HTTP server: %v", err)
		}
		log.Println("HTTP server stopped")
	}()

	return srv
}

func spinUpGRPCServer(grpcServer *grpc.Server, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("starting gRPC server on port %s", rpcAddr)
		lis, err := net.Listen("tcp", ":"+rpcAddr)
		if err != nil {
			log.Fatalf("failed to listen on %s: %v", rpcAddr, err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC server: %v", err)
		}
		log.Println("gRPC server stopped")
	}()
}

func stopHTTPServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}
}

func stopGRPCServer(grpcServer *grpc.Server) {
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-done:
		log.Println("gRPC server stopped gracefully")
	case <-ctx.Done():
		log.Println("gRPC server shutdown timed out, forcing stop")
		grpcServer.Stop()
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
		IsPrivate:       false,
	}

	s, err := storage.NewStorage(storageCfg)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}
	log.Println("minio connected")
	return s
}
