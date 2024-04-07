package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

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
		attrs:       logAttrsFromKeyAndValues(keyAndValues...),
		Enabled:     l.Enabled,
		spanContext: l.spanContext,
	}
}

func (l *Logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	sc := l.spanContext
	if sc == nil {
		sc = &spanContext{}
	}

	c, attrs, spanCtx := start(ctx, name, keyAndValues...)

	lgr := &Logger{
		Enabled: l.Enabled,

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

	l.spanContext.span.AddEvent(
		msgStr,
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(getProgressAttrs(l.attrs)...),
	)

	l.printf(l.spanContext.stdout, msgStr)
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

	span := l.spanContext.span
	span.RecordError(err, trace.WithAttributes(getProgressAttrs(l.attrs)...))
	span.SetStatus(codes.Error, errMsg)

	l.printf(l.spanContext.stderr, errMsg)
}

func (l *Logger) printf(w io.Writer, msg string) {
	if msg == "" {
		return
	}

	buf := bytes.NewBuffer(nil)
	defer func() {
		_, _ = io.Copy(w, buf)
	}()

	buf.WriteString(msg)

	for _, attr := range l.attrs {
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
