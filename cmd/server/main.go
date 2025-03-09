package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/config"
	"github.com/HellUpa/taskmanager/internal/db"
	"github.com/HellUpa/taskmanager/internal/http-server/handlers"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
)

// TODO: Изменить стандартный логгер на что нибудь более продвинутое. Например, slog.

func main() {
	// Configure the application.
	cfg := config.MustLoad()

	// Create a new meter provider.
	meterProvider, err := telemetry.NewPrometheusMeterProvider("taskmanager-server", "v0.1.0")
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
		log.Printf("failed to create request counter: %v", err)
	}
	requestLatency, err := telemetry.CreateHistogram(meter, "request_duration", "HTTP request duration (latency) in milliseconds", "ms")
	if err != nil {
		log.Printf("failed to create request latency histogram: %v", err)
	}

	// Connect to PostgreSQL.
	postgresDB, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer postgresDB.Close()

	// Create the TaskManager service.
	taskManagerService := app.NewTaskManagerService(postgresDB)

	// Create a new Chi router.
	r := chi.NewRouter()

	// Middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(telemetry.HTTPRequestMetrics(requestCount, requestLatency))

	// Routes.
	r.Get("/tasks", handlers.ListTasksHandler(taskManagerService))
	r.Post("/tasks", handlers.CreateTaskHandler(taskManagerService))
	r.Get("/tasks/{id}", handlers.GetTaskHandler(taskManagerService))
	r.Put("/tasks/{id}", handlers.UpdateTaskHandler(taskManagerService))
	r.Delete("/tasks/{id}", handlers.DeleteTaskHandler(taskManagerService))

	// Health check and metrics endpoints.
	h := chi.NewRouter()
	h.Get("/health", telemetry.HealthCheckHandler)
	m := chi.NewRouter()
	m.Handle("/metrics", telemetry.ExposeMetricsHandler())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine.
	server := &http.Server{
		Addr:         cfg.HTTPServer.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	go func() {
		fmt.Printf("Server listening on %v", cfg.HTTPServer.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start server: %v", err))
		}
	}()

	// Start Health Check server
	healthcheck := &http.Server{
		Addr:    cfg.HealthCheck.Port,
		Handler: h,
	}
	go func() {
		fmt.Printf("Healthcheck server listening on %v", cfg.HealthCheck.Port)
		if err := healthcheck.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start healthcheck server: %v", err))
		}
	}()

	// Start Metrics server
	metrics := &http.Server{
		Addr:    cfg.Telemetry.Port,
		Handler: m,
	}
	go func() {
		fmt.Printf("Metrics server listening on %v", cfg.Telemetry.Port)
		if err := metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start metrics server: %v", err))
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to gracefully shutdown server: %v", err)
	}

	if err := healthcheck.Shutdown(ctx); err != nil {
		log.Printf("failed to gracefully shutdown healthcheck server: %v", err)
	}

	if err := metrics.Shutdown(ctx); err != nil {
		log.Printf("failed to gracefully shutdown metrics server: %v", err)
	}
	log.Println("Server gracefully stopped")
}
