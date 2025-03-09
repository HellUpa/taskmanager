package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// MeterProvider интерфейс для провайдера метрик.
type MeterProvider interface {
	Shutdown(ctx context.Context) error
	Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter
}

// StdoutMeterProvider провайдер, выводящий метрики в stdout.
type StdoutMeterProvider struct {
	provider *sdkmetric.MeterProvider
}

// NewStdoutMeterProvider создает новый StdoutMeterProvider.
func NewStdoutMeterProvider(serviceName, serviceVersion string) (*StdoutMeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(5*time.Second))),
	)

	otel.SetMeterProvider(provider)

	return &StdoutMeterProvider{provider: provider}, nil
}

// Shutdown корректно завершает работу провайдера.
func (p *StdoutMeterProvider) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}

// Meter возвращает Meter для создания инструментов.
func (p *StdoutMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return p.provider.Meter(instrumentationName, opts...)
}

// CreateCounter создает и возвращает Int64Counter.
func CreateCounter(meter metric.Meter, name, description string) (metric.Int64Counter, error) {
	counter, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("failed to create counter %s: %w", name, err)
	}
	return counter, nil
}

// CreateHistogram creates a histogram to measure the distribution of request latencies.
func CreateHistogram(meter metric.Meter, name, description string, unit string) (metric.Int64Histogram, error) {
	histogram, err := meter.Int64Histogram(name,
		metric.WithDescription(description),
		metric.WithUnit(unit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram %s: %w", name, err)
	}
	return histogram, nil
}

// PrometheusMeterProvider провайдер, экспортирующий метрики в Prometheus.
type PrometheusMeterProvider struct {
	provider *sdkmetric.MeterProvider
}

// NewPrometheusMeterProvider создает новый PrometheusMeterProvider.
func NewPrometheusMeterProvider(serviceName, serviceVersion string) (*PrometheusMeterProvider, error) {

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	otel.SetMeterProvider(provider)

	return &PrometheusMeterProvider{provider: provider}, nil
}

// Shutdown ... (как и раньше)
func (p *PrometheusMeterProvider) Shutdown(ctx context.Context) error {
	return p.provider.Shutdown(ctx)
}

// Meter ... (как и раньше)
func (p *PrometheusMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return p.provider.Meter(instrumentationName, opts...)
}

// ExposeMetricsHandler возвращает http.Handler для эндпоинта /metrics.
func ExposeMetricsHandler() http.Handler {
	return promhttp.Handler()
}
