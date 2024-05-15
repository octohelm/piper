package logger

import (
	"context"

	"github.com/dagger/dagger/telemetry/sdklog"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Exporter interface {
	sdklog.LogExporter
	sdktrace.SpanExporter

	Run(context.Context, func(ctx context.Context) error) error
}
