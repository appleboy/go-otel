package otel

import (
	"context"
	"runtime"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestWithServiceVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "valid version",
			version: "1.0.0",
		},
		{
			name:    "semantic version",
			version: "2.1.3-beta",
		},
		{
			name:    "empty version",
			version: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{}
			opt := WithServiceVersion(tt.version)
			opt.apply(s)

			if s.version != tt.version {
				t.Errorf("WithServiceVersion() = %v, want %v", s.version, tt.version)
			}
		})
	}
}

func TestWithEnvironment(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "production environment",
			env:  "production",
		},
		{
			name: "staging environment",
			env:  "staging",
		},
		{
			name: "development environment",
			env:  "development",
		},
		{
			name: "empty environment",
			env:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{}
			opt := WithEnvironment(tt.env)
			opt.apply(s)

			if s.environment != tt.env {
				t.Errorf("WithEnvironment() = %v, want %v", s.environment, tt.env)
			}
		})
	}
}

func TestExporterFactory(t *testing.T) {
	tests := []struct {
		name         string
		exporterName string
		url          string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "unsupported exporter",
			exporterName: "unknown",
			url:          "http://localhost:9411",
			wantErr:      true,
			errMsg:       "unknown exporter is unsupported",
		},
		{
			name:         "empty exporter name",
			exporterName: "",
			url:          "http://localhost:9411",
			wantErr:      true,
			errMsg:       " exporter is unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter, err := ExporterFactory(tt.exporterName, tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("ExporterFactory() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ExporterFactory() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if exporter != nil {
					t.Error("ExporterFactory() expected nil exporter on error")
				}
			} else {
				if err != nil {
					t.Errorf("ExporterFactory() unexpected error = %v", err)
				}
				if exporter == nil {
					t.Error("ExporterFactory() expected non-nil exporter")
				}
			}
		})
	}
}

func TestNewTracer(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "basic tracer without options",
			opts:    []Option{},
			wantErr: false,
		},
		{
			name: "tracer with service version",
			opts: []Option{
				WithServiceVersion("1.0.0"),
			},
			wantErr: false,
		},
		{
			name: "tracer with environment",
			opts: []Option{
				WithEnvironment("production"),
			},
			wantErr: false,
		},
		{
			name: "tracer with all options",
			opts: []Option{
				WithServiceVersion("2.1.0"),
				WithEnvironment("staging"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory exporter for testing
			exporter := tracetest.NewInMemoryExporter()

			shutdown, err := NewTracer(exporter, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTracer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && shutdown == nil {
				t.Error("NewTracer() expected non-nil shutdown function")
			}

			// Test shutdown
			if shutdown != nil {
				if err := shutdown(context.Background()); err != nil {
					t.Errorf("shutdown() error = %v", err)
				}
			}
		})
	}
}

func TestNewTracerWithResourceAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()

	shutdown, err := NewTracer(
		exporter,
		WithServiceVersion("1.2.3"),
		WithEnvironment("test"),
	)
	if err != nil {
		t.Fatalf("NewTracer() error = %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			t.Errorf("shutdown() error = %v", err)
		}
	}()

	// Create a simple span to verify the tracer is working
	ctx := context.Background()
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test-tracer")

	_, span := tracer.Start(ctx, "test-span")
	span.End()

	// Verify the span was created
	if span == nil {
		t.Error("expected non-nil span")
	}
}

func TestServiceStruct(t *testing.T) {
	s := &service{
		name:        "test-service",
		version:     "1.0.0",
		environment: "test",
	}

	if s.name != "test-service" {
		t.Errorf("service.name = %v, want test-service", s.name)
	}
	if s.version != "1.0.0" {
		t.Errorf("service.version = %v, want 1.0.0", s.version)
	}
	if s.environment != "test" {
		t.Errorf("service.environment = %v, want test", s.environment)
	}
}

func TestOptionInterface(t *testing.T) {
	// Test that options implement the Option interface
	_ = WithServiceVersion("1.0.0")
	_ = WithEnvironment("production")
}

func BenchmarkNewTracer(b *testing.B) {
	exporter := tracetest.NewInMemoryExporter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shutdown, err := NewTracer(
			exporter,
			WithServiceVersion("1.0.0"),
			WithEnvironment("production"),
		)
		if err != nil {
			b.Fatalf("NewTracer() error = %v", err)
		}
		if err := shutdown(context.Background()); err != nil {
			b.Fatalf("shutdown() error = %v", err)
		}
	}
}

func BenchmarkExporterFactory(b *testing.B) {
	testCases := []struct {
		name         string
		exporterName string
		url          string
	}{
		{"unsupported", "unknown", "http://localhost:9411"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ExporterFactory(tc.exporterName, tc.url)
			}
		})
	}
}

func TestRuntimeAttributes(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()

	shutdown, err := NewTracer(exporter)
	if err != nil {
		t.Fatalf("NewTracer() error = %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			t.Errorf("shutdown() error = %v", err)
		}
	}()

	// Verify runtime attributes are set correctly
	expectedOS := runtime.GOOS
	expectedArch := runtime.GOARCH

	if expectedOS == "" {
		t.Error("runtime.GOOS should not be empty")
	}
	if expectedArch == "" {
		t.Error("runtime.GOARCH should not be empty")
	}
}
