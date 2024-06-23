package logger

import (
	"context"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Exporter interface {
	sdklog.Exporter
	sdktrace.SpanExporter

	Run(context.Context, func(ctx context.Context) error) error
}
