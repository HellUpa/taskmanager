package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

// HTTPRequestMetrics is a middleware to count HTTP requests and record their duration.
func HTTPRequestMetrics(counter metric.Int64Counter, histogram metric.Int64Histogram) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Record the start time.
			startTime := time.Now()

			// This is just a placeholder. You'll get the actual user ID from
			// your authentication middleware (Ory Kratos/Hydra integration).
			userID := uuid.New() // Generate a new user ID from context.

			// Add the user ID to the context.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// Call the next handler in the chain.
			next.ServeHTTP(w, r.WithContext(ctx))

			// Record the duration (latency).
			duration := time.Since(startTime).Milliseconds()
			histogram.Record(r.Context(), duration, metric.WithAttributes(attribute.String("method", r.Method), attribute.String("path", r.URL.Path)))

			// Increment the counter for each request.
			counter.Add(r.Context(), 1, metric.WithAttributes(attribute.String("method", r.Method), attribute.String("path", r.URL.Path)))
		})
	}
}
