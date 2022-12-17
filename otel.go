package otel

import (
	"context"
	"fmt"
	"runtime"

	"github.com/appleboy/go-otel/zipkin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerProvider interface {
	Tracer(...trace.TracerOption) trace.Tracer
	Shutdown(context.Context) error
}

type Option interface {
	apply(cfg *service)
}

type option func(cfg *service)

func (fn option) apply(cfg *service) {
	fn(cfg)
}

// WithServiceVersion configures `service.version` resource attribute, for example, `1.0.0`.
func WithServiceVersion(serviceVersion string) Option {
	return option(func(cfg *service) {
		cfg.version = serviceVersion
	})
}

// WithDeploymentEnvironment configures `deployment.environment` resource attribute,
// for example, `production`.
func WithEnvironment(env string) Option {
	return option(func(cfg *service) {
		cfg.environment = env
	})
}

type service struct {
	name        string
	version     string
	environment string
}

// ExporterFactory - Create tracer according to given params
func ExporterFactory(name, url string) (sdktrace.SpanExporter, error) {
	switch name {
	case "zipkin":
		return zipkin.New(url)
	default:
		return nil, fmt.Errorf("%s exporter is unsupported", name)
	}
}

// NewTracer - Creates new tracer
func NewTracer(exporter sdktrace.SpanExporter, opts ...Option) (func(context.Context) error, error) {
	s := &service{}

	for _, opt := range opts {
		opt.apply(s)
	}

	resources, err := resource.New(
		context.Background(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(s.name),
			semconv.ServiceVersionKey.String(s.version),
			attribute.String("service.language", "go"),
			attribute.String("service.environment", s.environment),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
