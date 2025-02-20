package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/HellUpa/gRPC-CRUD/internal/app"
	"github.com/HellUpa/gRPC-CRUD/internal/db"
	pb "github.com/HellUpa/gRPC-CRUD/pb/gen"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// Database connection parameters (use environment variables).
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Connect to PostgreSQL.
	postgresDB, err := db.NewPostgresDB(dbHost, dbPort, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgresDB.Close()

	// Create the TaskManager service.
	taskManagerService := app.NewTaskManagerService(postgresDB)

	// Create a gRPC server.
	grpcServer := grpc.NewServer()

	// Register the TaskManager service with the gRPC server.
	pb.RegisterTaskManagerServer(grpcServer, taskManagerService)

	// Listen on a port (e.g., 50051).
	port := "50051"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Server listening on port %s\n", port)
	// Start the gRPC server.
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
