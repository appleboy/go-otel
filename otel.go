package otel

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type TracerProvider interface {
	Tracer(...trace.TracerOption) trace.Tracer
	Shutdown(context.Context) error
}
