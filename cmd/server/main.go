package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/HellUpa/gRPC-CRUD/internal/app"
	"github.com/HellUpa/gRPC-CRUD/internal/db"
	pb "github.com/HellUpa/gRPC-CRUD/pb/gen"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Database connection parameters (use flags).
	var dbHost, dbPort, dbUser, dbPassword, dbName string

	flag.StringVar(&dbHost, "host", "localhost", "db address")
	flag.StringVar(&dbPort, "port", "5432", "db port")
	flag.StringVar(&dbUser, "user", "postgres", "db user")
	flag.StringVar(&dbPassword, "password", "postgres", "db password")
	flag.StringVar(&dbName, "db_name", "postgres", "db name")
	flag.Parse()

	// Connect to PostgreSQL.
	postgresDB, err := db.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
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

	reflection.Register(grpcServer)

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
