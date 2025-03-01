package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
)

// UnaryInterceptor возвращает унарный interceptor, который увеличивает счетчик запросов.
func UnaryInterceptor(counter metric.Int64Counter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Увеличиваем счетчик запросов.
		counter.Add(ctx, 1)

		// Вызываем следующий обработчик в цепочке.
		return handler(ctx, req)
	}
}
