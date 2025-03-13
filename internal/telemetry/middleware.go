package telemetry

import (
	"net/http"
	"time"

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

			// Call the next handler in the chain.
			next.ServeHTTP(w, r)

			// Record the duration (latency).
			duration := time.Since(startTime).Milliseconds()
			histogram.Record(r.Context(), duration, metric.WithAttributes(attribute.String("method", r.Method), attribute.String("path", r.URL.Path)))

			// Increment the counter for each request.
			counter.Add(r.Context(), 1, metric.WithAttributes(attribute.String("method", r.Method), attribute.String("path", r.URL.Path)))
		})
	}
}
