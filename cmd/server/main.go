package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/config"
	"github.com/HellUpa/taskmanager/internal/db"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	pb "github.com/HellUpa/taskmanager/pb/gen"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// TODO: Изменить стандартный логгер на что нибудь более продвинутое. Например, slog.

func main() {
	// Configure the application.
	cfg := config.MustLoad()

	// Create a new meter provider.
	meterProvider, err := telemetry.NewStdoutMeterProvider("taskmanager-server", "v0.1.0")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	meter := meterProvider.Meter("taskmanager-server")
	requestCount, err := telemetry.CreateCounter(meter, "requests_total", "Total number of requests")
	if err != nil {
		log.Fatalf("failed to create request counter: %v", err)
	}

	// Connect to PostgreSQL.
	postgresDB, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgresDB.Close()

	// Create the TaskManager service.
	taskManagerService := app.NewTaskManagerService(postgresDB)

	// Create a new gRPC server. We also attach the unary interceptor to the server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(telemetry.UnaryInterceptor(requestCount)),
	)

	// Register the TaskManager service with the gRPC server.
	pb.RegisterTaskManagerServer(grpcServer, taskManagerService)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	go func() {
		http.HandleFunc("/health", telemetry.HealthCheckHandler)
		log.Printf("Health check server listening on :%d\n", cfg.Telemetry.HealthCheckPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Telemetry.HealthCheckPort), nil); err != nil {
			log.Fatalf("failed to start health check server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Server listening on port %s\n", cfg.GRPC.Port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the gRPC server in a goroutine.
	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server gracefully stopped")
}
