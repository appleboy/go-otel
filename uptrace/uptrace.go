package uptrace

import (
	"context"

	core "github.com/appleboy/go-otel"

	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel"
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
	name        string
	version     string
	environment string
}

func (s *service) Tracer(opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(s.name, opts...)
}

func (s *service) Shutdown(ctx context.Context) error {
	return uptrace.Shutdown(ctx)
}

func (s *service) Apply(opts ...uptrace.Option) core.TracerProvider {
	options := append(
		[]uptrace.Option{},
		uptrace.WithServiceName(s.name),
		uptrace.WithServiceVersion(s.version),
		uptrace.WithDeploymentEnvironment(s.environment),
	)

	options = append(options, opts...)

	uptrace.ConfigureOpentelemetry(
		options...,
	)
	return s
}

func New(
	name string,
	opts ...Option,
) core.TracerProvider {
	s := &service{
		name: name,
	}

	for _, opt := range opts {
		opt.apply(s)
	}

	uptrace.ConfigureOpentelemetry(
		uptrace.WithServiceName(s.name),
		uptrace.WithServiceVersion(s.version),
		uptrace.WithDeploymentEnvironment(s.environment),
	)

	return s
}
