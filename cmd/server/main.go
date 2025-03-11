package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellUpa/taskmanager/internal/app"
	"github.com/HellUpa/taskmanager/internal/config"
	"github.com/HellUpa/taskmanager/internal/db"
	"github.com/HellUpa/taskmanager/internal/http-server/handlers"
	"github.com/HellUpa/taskmanager/internal/logger"
	logu "github.com/HellUpa/taskmanager/internal/logger/logger-utils"
	"github.com/HellUpa/taskmanager/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
)

func main() {
	// Configure the application.
	cfg := config.MustLoad()
	// Setup the logger.
	log := logger.SetupLogger(cfg.Env)

	// Create a new meter provider.
	meterProvider, err := telemetry.NewPrometheusMeterProvider("taskmanager-server", "v0.1.0")
	if err != nil {
		log.Error("Error starting meter provider", logu.Err(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Error("Error shutting down meter provider", logu.Err(err))
		}
	}()

	meter := meterProvider.Meter("taskmanager-server")
	requestCount, err := telemetry.CreateCounter(meter, "requests_total", "Total number of requests")
	if err != nil {
		log.Error("Failed to create request counter", logu.Err(err))
	}
	requestLatency, err := telemetry.CreateHistogram(meter, "request_duration", "HTTP request duration (latency) in milliseconds", "ms")
	if err != nil {
		log.Error("Failed to create request latency histogram", logu.Err(err))
	}

	// Connect to PostgreSQL.
	postgresDB, err := db.NewPostgresDB(log, cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database", logu.Err(err))
	}
	defer postgresDB.Close()

	// Create the TaskManager service.
	taskManagerService := app.NewTaskManagerService(log, postgresDB)

	// Create a new Chi router.
	r := chi.NewRouter()

	// Middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.NewMiddlewareLogger(log))
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
		log.Info("Starting server", slog.String("port", cfg.HTTPServer.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", logu.Err(err))
			os.Exit(1)
		}
	}()

	// Start Healthcheck server
	healthcheck := &http.Server{
		Addr:    cfg.HealthCheck.Port,
		Handler: h,
	}
	go func() {
		log.Info("Starting healthcheck", slog.String("port", cfg.HealthCheck.Port))
		if err := healthcheck.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start healthcheck server", logu.Err(err))
			os.Exit(1)
		}
	}()

	// Start Metrics server
	metrics := &http.Server{
		Addr:    cfg.Telemetry.Port,
		Handler: m,
	}
	go func() {
		log.Info("Starting metrics", slog.String("port", cfg.Telemetry.Port))
		if err := metrics.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start metrics server", logu.Err(err))
			os.Exit(1)
		}
	}()

	<-stop
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Failed to gracefully shutdown server", logu.Err(err))
	}

	if err := healthcheck.Shutdown(ctx); err != nil {
		log.Error("Failed to gracefully shutdown healthcheck server", logu.Err(err))
	}

	if err := metrics.Shutdown(ctx); err != nil {
		log.Error("Failed to gracefully shutdown metrics server", logu.Err(err))
	}
	log.Info("Server gracefully stopped")
}
