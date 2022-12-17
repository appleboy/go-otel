package jaeger

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
)

// New - Creates new Jaeger exporter
func New(url string) (trace.SpanExporter, error) {
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)),
	)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}
