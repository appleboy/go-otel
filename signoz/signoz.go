package signoz

import (
	"context"
	"io"
	"runtime"
	"strings"

	core "github.com/appleboy/go-otel"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

type Option interface {
	apply(cfg *service)
}

type option func(cfg *service)

func (fn option) apply(cfg *service) {
	fn(cfg)
}

// WithServiceVersion configures `service.version` resource attribute, for example, `1.0.0`.
func WithVersion(serviceVersion string) Option {
	return option(func(s *service) {
		s.version = serviceVersion
	})
}

// WithEnvironment configures `service.environment` resource attribute,
// for example, `production`.
func WithEnvironment(env string) Option {
	return option(func(s *service) {
		s.environment = env
	})
}

func WithHeaders(envs map[string]string) Option {
	return option(func(s *service) {
		s.headers = envs
	})
}

func WithCollectorURL(url string) Option {
	return option(func(s *service) {
		s.collectorURL = url
	})
}

type service struct {
	name         string
	version      string
	environment  string
	collectorURL string
	headers      map[string]string
	tp           *sdktrace.TracerProvider
}

func (s *service) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(s.name, opts...)
}

func (s *service) Shutdown(ctx context.Context) error {
	return s.tp.Shutdown(ctx)
}

func New(
	name string,
	opts ...Option,
) (core.TracerProvider, error) {
	s := &service{
		name: name,
	}

	for _, opt := range opts {
		opt.apply(s)
	}

	var exporter sdktrace.SpanExporter
	var err error

	if s.collectorURL != "" {
		headers := map[string]string{}

		secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if s.insecure() {
			secureOption = otlptracegrpc.WithInsecure()
		}

		if s.headers != nil {
			for k, v := range s.headers {
				headers[k] = v
			}
		}

		exporter, err = otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(
				secureOption,
				otlptracegrpc.WithEndpoint(s.collectorURL),
				otlptracegrpc.WithHeaders(headers),
			),
		)
		if err != nil {
			return nil, err
		}
	} else {
		exporter, err = stdout.New(
			stdout.WithWriter(io.Discard),
			stdout.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", s.name),
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

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	s.tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(s.tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return s, nil
}

func (s *service) insecure() bool {
	return !strings.HasPrefix(s.collectorURL, "https://")
}
