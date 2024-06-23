package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"dagger.io/dagger/telemetry"
	"github.com/octohelm/piper/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

const LogAttrSpanName = "$$spanName"

type spanContext struct {
	ctx       context.Context
	span      trace.Span
	name      string
	startedAt time.Time
	log       log.Logger
}

func start(ctx context.Context, spanName string, keyAndValues ...any) (context.Context, []slog.Attr, *spanContext) {
	spanCtx := &spanContext{}
	spanCtx.name = spanName
	spanCtx.startedAt = time.Now()

	attrs := logAttrsFromKeyAndValues(keyAndValues...)

	t := TracerContext.From(ctx)

	c, span := t.Start(ctx, spanCtx.name, trace.WithTimestamp(spanCtx.startedAt), trace.WithAttributes(getProgressAttrs(attrs)...))

	spanCtx.span = span
	spanCtx.ctx = c
	spanCtx.log = telemetry.Logger(spanName)

	return c, attrs, spanCtx
}

func getProgressAttrs(attrs []slog.Attr) []attribute.KeyValue {
	finalAttrs := make([]attribute.KeyValue, len(attrs))

	for _, attr := range attrs {
		switch attr.Key {
		case otel.LogAttrProgressCurrent:
			finalAttrs = append(
				finalAttrs,
				attribute.Int64(telemetry.ProgressCurrentAttr, attr.Value.Int64()),
			)
		case otel.LogAttrProgressTotal:
			finalAttrs = append(
				finalAttrs,
				attribute.Int64(telemetry.ProgressTotalAttr, attr.Value.Int64()),
			)
		}
	}

	return finalAttrs
}

func logAttrsFromKeyAndValues(keysAndValues ...any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(keysAndValues))

	i := 0

	for i < len(keysAndValues) {
		switch x := keysAndValues[i].(type) {
		case slog.Attr:
			attrs = append(attrs, x)
			i++
		case attribute.KeyValue:
			attrs = append(attrs, slog.Any(string(x.Key), x.Value.AsInterface()))
			i++
		case string:
			key := x
			// x
			if i+1 < len(keysAndValues) {
				// get value
				i++
				v := keysAndValues[i]

				switch x := v.(type) {
				case slog.LogValuer:
					attrs = append(attrs, slog.Any(key, x))
				case []byte:
					attrs = append(attrs, slog.Any(key, x))
				case fmt.Stringer:
					attrs = append(attrs, slog.String(key, x.String()))
				case string:
					attrs = append(attrs, slog.String(key, x))
				case int:
					attrs = append(attrs, slog.Int(key, x))
				case float64:
					attrs = append(attrs, slog.Float64(key, x))
				case bool:
					attrs = append(attrs, slog.Bool(key, x))
				default:
					attrs = append(attrs, slog.String(key, fmt.Sprintf("%v", x)))
				}
			}
			i++
		default:
			fmt.Printf("bad attr %#v", x)
			i++
		}
	}

	return attrs
}

func sprintf(format string, args ...any) fmt.Stringer {
	return &printer{format: format, args: args}
}

type printer struct {
	format string
	args   []any
}

func (p *printer) String() string {
	if len(p.args) == 0 {
		return p.format
	}
	return fmt.Sprintf(p.format, p.args...)
}
