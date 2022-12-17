package signoz

import (
	"context"
	"io"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type Option interface {
	apply(cfg *service)
}

type option func(cfg *service)

func (fn option) apply(cfg *service) {
	fn(cfg)
}

func WithHeaders(envs map[string]string) Option {
	return option(func(s *service) {
		s.headers = envs
	})
}

type service struct {
	headers map[string]string
}

func New(
	url string,
	opts ...Option,
) (trace.SpanExporter, error) {
	s := &service{}

	for _, opt := range opts {
		opt.apply(s)
	}

	var exporter trace.SpanExporter
	var err error

	if url != "" {
		headers := map[string]string{}

		secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if insecure(url) {
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
				otlptracegrpc.WithEndpoint(url),
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

	return exporter, nil
}

func insecure(url string) bool {
	return !strings.HasPrefix(url, "https://")
}
