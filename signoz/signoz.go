package signoz

import (
	"context"
	"io"

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

type Service struct {
	name string
	tp   *sdktrace.TracerProvider
}

func (s *Service) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(s.name, opts...)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return s.tp.Shutdown(ctx)
}

func New(
	serviceName string,
	collectorURL string,
	insecure bool,
) (core.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if collectorURL != "" {
		headers := map[string]string{}

		secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if insecure {
			secureOption = otlptracegrpc.WithInsecure()
		}

		exporter, err = otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(
				secureOption,
				otlptracegrpc.WithEndpoint(collectorURL),
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
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return &Service{
		tp:   traceProvider,
		name: serviceName,
	}, nil
}
