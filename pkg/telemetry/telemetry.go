package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc/credentials/insecure"
)

// InitTracer는 Jaeger(OTLP)로 Trace를 보내는 Provider를 설정합니다.
func InitTracer(serviceName string, collectorAddr string) func(context.Context) error {
	ctx := context.Background()

	// 1. OTLP gRPC Exporter 설정 (Jaeger 수집기 주소)
	// collectorAddr 예시: "jaeger.monitoring.svc.cluster.local:4317"
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(collectorAddr),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	// 2. Resource 설정 (서비스 이름 등)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// 3. TracerProvider 생성
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// 4. 전역 Tracer 및 Propagator 설정
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown
}
