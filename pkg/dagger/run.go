package dagger

import (
	"context"

	"dagger.io/dagger/telemetry"
	"github.com/dagger/dagger/dagql/dagui"
	"github.com/dagger/dagger/engine/slog"
	"github.com/dagger/dagger/util/cleanups"
	"go.opentelemetry.io/otel"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/octohelm/piper/internal/logger"
	"github.com/octohelm/piper/internal/version"
)

func Run(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	frontend := logger.NewFrontend()

	run := func(ctx context.Context) (rerr error) {
		defer telemetry.Close()

		ctx = telemetry.Init(ctx, telemetry.Config{
			Resource: resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("piper"),
				semconv.ServiceVersion(version.Version()),
			),
			LiveTraceExporters: []sdktrace.SpanExporter{frontend.SpanExporter()},
			LiveLogExporters:   []sdklog.Exporter{frontend.LogExporter()},
		})

		daggerRunner, err := NewRunner(
			WithCloudURLCallback(frontend.SetCloudURL),
			WithEngineLogs(frontend.LogExporter()),
			WithEngineTrace(frontend.SpanExporter()),
		)
		if err != nil {
			return err
		}

		tracer := Tracer()

		ctx = logger.TracerContext.Inject(ctx, tracer)

		c, span := tracer.Start(ctx, name)
		defer telemetry.End(span, func() error { return rerr })

		// important for exec logging
		frontend.SetPrimary(dagui.SpanID{
			SpanID: span.SpanContext().SpanID(),
		})
		slog.SetDefault(slog.SpanLogger(ctx, name))

		return fn(RunnerContext.Inject(c, daggerRunner))
	}

	return frontend.Run(ctx, dagui.FrontendOpts{TooFastThreshold: 1}, func(ctx context.Context) (cleanups.CleanupF, error) {
		return func() error { return nil }, run(ctx)
	})
}

func Tracer() trace.Tracer {
	return otel.Tracer("piper.octohelm.tech/cli")
}
