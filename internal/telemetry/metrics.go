package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
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

// TODO: Сделать нормальный провайдер метрик.(Например, PrometheusMeterProvider)

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

	otel.SetMeterProvider(provider) // Устанавливаем глобальный провайдер

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
