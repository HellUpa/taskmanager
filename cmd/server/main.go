package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellUpa/gRPC-CRUD/internal/app"
	"github.com/HellUpa/gRPC-CRUD/internal/db"
	"github.com/HellUpa/gRPC-CRUD/internal/telemetry"
	pb "github.com/HellUpa/gRPC-CRUD/pb/gen"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// TODO: Изменить стандартный логгер на что нибудь более продвинутое. Например, slog.

func main() {
	// Database connection parameters (use flags).
	var dbHost, dbPort, dbUser, dbPassword, dbName, listenPort string
	var healthCheckPort int

	flag.StringVar(&dbHost, "host", "localhost", "db address")
	flag.StringVar(&dbPort, "port", "5432", "db port")
	flag.StringVar(&dbUser, "user", "postgres", "db user")
	flag.StringVar(&dbPassword, "password", "postgres", "db password")
	flag.StringVar(&dbName, "dbname", "taskmanager", "db name")
	flag.StringVar(&listenPort, "listen", "50051", "gRPC server listen port")
	flag.IntVar(&healthCheckPort, "healthcheck", 8080, "Health check port")
	flag.Parse()

	// Инициализируем провайдер метрик.
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
	postgresDB, err := db.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgresDB.Close()

	// Create the TaskManager service.
	taskManagerService := app.NewTaskManagerService(postgresDB)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(telemetry.UnaryInterceptor(requestCount)),
	)

	// Register the TaskManager service with the gRPC server.
	pb.RegisterTaskManagerServer(grpcServer, taskManagerService)

	reflection.Register(grpcServer)

	go func() {
		http.HandleFunc("/healthz", telemetry.HealthCheckHandler)
		log.Printf("Health check server listening on :%d\n", healthCheckPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", healthCheckPort), nil); err != nil {
			log.Fatalf("failed to start health check server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Server listening on port %s\n", listenPort)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем gRPC-сервер в отдельной горутине.
	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	// Ожидаем сигнал.
	<-stop
	log.Println("Shutting down server...")
	// Graceful shutdown gRPC-сервера.
	grpcServer.GracefulStop()
	log.Println("Server gracefully stopped")
}
