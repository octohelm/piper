package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/octohelm/piper/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"

	"github.com/fatih/color"
	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	Enabled     logr.Level
	attrs       []slog.Attr
	spanContext *spanContext
}

func (l *Logger) WithValues(keyAndValues ...any) logr.Logger {
	return &Logger{
		Enabled:     l.Enabled,
		spanContext: l.spanContext,
		attrs:       append(l.attrs, logAttrsFromKeyAndValues(keyAndValues...)...),
	}
}

func (l *Logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	sc := l.spanContext
	if sc == nil {
		sc = &spanContext{
			ctx: ctx,
		}
	}

	c, attrs, spanCtx := start(ctx, name, keyAndValues...)

	lgr := &Logger{
		Enabled:     l.Enabled,
		attrs:       append(l.attrs, attrs...),
		spanContext: spanCtx,
	}

	return logr.WithLogger(c, lgr), lgr
}

func (l *Logger) End() {
	if l.spanContext == nil {
		return
	}
	l.spanContext.span.End(trace.WithTimestamp(time.Now()))
}

func (l *Logger) Debug(format string, args ...any) {
	l.info(logr.DebugLevel, sprintf(format, args...))
}

func (l *Logger) Info(format string, args ...any) {
	l.info(logr.InfoLevel, sprintf(format, args...))
}

func (l *Logger) Warn(err error) {
	l.error(logr.WarnLevel, err)
}

func (l *Logger) Error(err error) {
	l.error(logr.ErrorLevel, err)
}

func (l *Logger) info(level logr.Level, msg fmt.Stringer) {
	if l.spanContext == nil {
		return
	}

	if level > l.Enabled {
		return
	}

	msgStr := msg.String()

	keyValues := getProgressAttrs(l.attrs)

	l.spanContext.span.AddEvent(
		msgStr,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(keyValues...),
	)

	l.printf(l.spanContext.log, msgStr, keyValues)
}

func (l *Logger) error(level logr.Level, err error) {
	if l.spanContext == nil {
		return
	}

	if level > l.Enabled {
		return
	}

	if err == nil {
		return
	}

	errMsg := err.Error()
	keyValues := getProgressAttrs(l.attrs)

	span := l.spanContext.span
	span.RecordError(err, trace.WithAttributes(keyValues...))
	span.SetStatus(codes.Error, errMsg)

	l.printf(l.spanContext.log, errMsg, keyValues)
}

func (l *Logger) printf(ll log.Logger, msg string, attrs []attribute.KeyValue) {
	if msg == "" {
		return
	}

	rec := log.Record{}
	rec.SetTimestamp(time.Now())
	rec.AddAttributes(log.String(LogAttrSpanName, l.spanContext.name))

	for _, attr := range attrs {
		rec.AddAttributes(log.KeyValue{Key: string(attr.Key), Value: log.Int64Value(attr.Value.AsInt64())})
	}

	buf := bytes.NewBuffer(nil)
	defer func() {
		rec.SetBody(log.StringValue(buf.String()))
		ll.Emit(l.spanContext.ctx, rec)
	}()

	buf.WriteString(msg)

	for _, attr := range l.attrs {
		switch attr.Key {
		case otel.LogAttrProgressCurrent, otel.LogAttrProgressTotal:
			continue
		}

		switch x := attr.Value.Any().(type) {
		case string:
			_, _ = fmt.Fprintf(buf, color.WhiteString(" %s=%q", attr.Key, attr.Value))
		default:
			logValue := x
			if valuer, ok := logValue.(slog.LogValuer); ok {
				logValue = valuer.LogValue().Any()
			}

			switch x := logValue.(type) {
			case []byte:
				_, _ = fmt.Fprint(buf, color.WhiteString(" %s=", attr.Key))
				s := bufio.NewScanner(bytes.NewBuffer(x))
				for s.Scan() {
					if line := s.Text(); len(line) > 0 {
						_, _ = fmt.Fprint(buf, color.WhiteString("%s\n", line))
					}
				}
			default:
				_, _ = fmt.Fprint(buf, color.WhiteString(" %s=%v", attr.Key, x))
			}
		}
	}

	buf.WriteString("\n")
}
