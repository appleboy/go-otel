package uptrace

import (
	"context"

	core "github.com/appleboy/go-otel"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	name string
}

func (s *Service) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(s.name, opts...)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return uptrace.Shutdown(ctx)
}

func New(
	serviceName string,
) (core.TracerProvider, error) {
	uptrace.ConfigureOpentelemetry(
		// copy your project DSN here or use UPTRACE_DSN env var
		// uptrace.WithDSN("https://<token>@uptrace.dev/<project_id>"),

		uptrace.WithServiceName(serviceName),
		uptrace.WithServiceVersion("1.0.0"),
	)

	return &Service{
		name: serviceName,
	}, nil
}
