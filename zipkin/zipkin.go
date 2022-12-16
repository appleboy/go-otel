package zipkin

import (
	"context"
	"runtime"

	core "github.com/appleboy/go-otel"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

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
	url         string
	name        string
	version     string
	environment string

	tp *sdktrace.TracerProvider
}

func (s *service) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(s.name, opts...)
}

func (s *service) Shutdown(ctx context.Context) error {
	return s.tp.Shutdown(ctx)
}

// NewZipkin - Creates new Zipkin exporter
func New(url string, opts ...Option) (core.TracerProvider, error) {
	s := &service{
		url: url,
	}

	for _, opt := range opts {
		opt.apply(s)
	}

	exporter, err := zipkin.New(url)
	if err != nil {
		return nil, err
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(s.name),
			attribute.String("service.language", "go"),
			attribute.String("service.version", s.version),
			attribute.String("service.environment", s.environment),
			attribute.String("os", runtime.GOOS),
			attribute.String("arch", runtime.GOARCH),
		),
	)
	if err != nil {
		return nil, err
	}

	s.tp = sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(s.tp)
	return s, nil
}
