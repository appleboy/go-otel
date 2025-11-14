package jaeger

import (
	"go.opentelemetry.io/otel/exporters/jaeger" //nolint:staticcheck // Keeping for backward compatibility
	"go.opentelemetry.io/otel/sdk/trace"
)

// New - Creates new Jaeger exporter
//
// Deprecated: The Jaeger exporter is deprecated. OpenTelemetry has dropped support for the
// Jaeger exporter. Jaeger officially accepts and recommends using OTLP instead.
// Please use OTLP exporters (otlptracehttp or otlptracegrpc) instead.
func New(url string) (trace.SpanExporter, error) {
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
