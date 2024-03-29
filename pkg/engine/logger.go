package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/opencontainers/go-digest"
	"github.com/vito/progrock"
	"io"
	"log/slog"
	"strings"
)

// +gengo:enum
type LogLevel string

const (
	ErrorLevel LogLevel = "error"
	WarnLevel  LogLevel = "warn"
	InfoLevel  LogLevel = "info"
	DebugLevel LogLevel = "debug"
)

type Logger struct {
	// Log level
	LogLevel LogLevel `flag:",omitempty"`

	logger logr.Logger
}

func (l *Logger) SetDefaults() {
	if l.LogLevel == "" {
		l.LogLevel = InfoLevel
	}
}

func (l *Logger) Init(ctx context.Context) error {
	if l.logger == nil {
		lvl, _ := logr.ParseLevel(string(l.LogLevel))
		l.logger = &logger{
			lvl: lvl,
			r:   progrock.FromContext(ctx).Vertex(digest.FromString(progrock.RootID), progrock.RootID),
		}
	}
	return nil
}

func (l *Logger) InjectContext(ctx context.Context) context.Context {
	return configuration.InjectContext(
		ctx,
		configuration.InjectContextFunc(logr.WithLogger, l.logger),
	)
}

type logger struct {
	lvl   logr.Level
	attrs []slog.Attr
	r     Recoder
}

type Recoder interface {
	Stdout() io.Writer
	Stderr() io.Writer
}

func (l *logger) WithValues(keyAndValues ...any) logr.Logger {
	attrs := append(l.attrs, argsToAttrSlice(keyAndValues)...)

	return &logger{
		lvl:   l.lvl,
		r:     l.r,
		attrs: attrs,
	}
}

func (l *logger) Start(ctx context.Context, name string, keyAndValues ...any) (context.Context, logr.Logger) {
	rec := progrock.FromContext(ctx)
	display := name
	attrs := append(l.attrs, argsToAttrSlice(keyAndValues)...)
	finalAttrs := make([]slog.Attr, 0, len(attrs))
	inputs := make([]digest.Digest, 0)
	taskTotal := int64(0)

	for _, attr := range attrs {
		switch attr.Key {
		case cueflow.LogAttrProgressTotal:
			taskTotal = attr.Value.Int64()
		case cueflow.LogAttrName:
			display = color.WhiteString(attr.Value.String())
			continue
		case cueflow.LogAttrDep:
			inputs = append(inputs, digest.FromString(attr.Value.String()))
			continue
		case cueflow.LogAttrScope:
			rec = rec.WithGroup(attr.Value.String(), progrock.WithGroupID(attr.Value.String()))
			continue
		}
		finalAttrs = append(finalAttrs, attr)
	}

	if taskTotal > 0 {
		if vtx, ok := l.r.(*progrock.VertexRecorder); ok {
			task := vtx.ProgressTask(taskTotal, display)
			task.Start()

			cl := &logger{
				lvl:   l.lvl,
				r:     task,
				attrs: finalAttrs,
			}

			ctx = progrock.ToContext(ctx, rec)
			ctx = logr.WithLogger(ctx, cl)

			return ctx, cl
		}
	}

	r := rec.WithGroup(name, progrock.WithGroupID(name), progrock.Weak()).Vertex(
		digest.FromString(name),
		display,
		progrock.WithInputs(inputs...),
	)

	cl := &logger{
		lvl:   l.lvl,
		r:     r,
		attrs: finalAttrs,
	}

	ctx = progrock.ToContext(ctx, rec)
	ctx = logr.WithLogger(ctx, cl)

	return ctx, cl
}

func (l *logger) End() {
	switch x := l.r.(type) {
	case *progrock.VertexRecorder:
		x.Done(nil)
	case *progrock.TaskRecorder:
		x.Complete()
	}

	l.attrs = nil
}

func (l *logger) Enabled(lvl logr.Level) bool {
	return lvl <= l.lvl
}

func (l *logger) Debug(format string, args ...any) {
	if !l.Enabled(logr.DebugLevel) {
		return
	}

	if len(args) > 0 {
		l.printf(l.r.Stdout(), color.WhiteString(strings.ToUpper(logr.DebugLevel.String()[:4])), fmt.Sprintf(format, args...))
	} else {
		l.printf(l.r.Stdout(), color.WhiteString(strings.ToUpper(logr.DebugLevel.String()[:4])), format)
	}
}

func (l *logger) Info(format string, args ...any) {
	if !l.Enabled(logr.InfoLevel) {
		return
	}

	if len(args) > 0 {
		l.printf(l.r.Stdout(), color.GreenString(strings.ToUpper(logr.InfoLevel.String()[:4])), fmt.Sprintf(format, args...))
	} else {
		l.printf(l.r.Stdout(), color.GreenString(strings.ToUpper(logr.InfoLevel.String()[:4])), format)
	}
}

func (l *logger) Warn(err error) {
	if err == nil || !l.Enabled(logr.WarnLevel) {
		return
	}

	l.printf(l.r.Stderr(), color.YellowString(strings.ToUpper(logr.InfoLevel.String()[:4])), err.Error())
}

func (l *logger) Error(err error) {
	if err == nil || !l.Enabled(logr.ErrorLevel) {
		return
	}

	l.printf(l.r.Stderr(), color.RedString(strings.ToUpper(logr.ErrorLevel.String()[:4])), err.Error())
}

func (l *logger) printf(w io.Writer, prefix string, msg string) {
	for _, attr := range l.attrs {
		if attr.Key == cueflow.LogAttrProgressCurrent {
			switch x := l.r.(type) {
			case *progrock.TaskRecorder:
				x.Current(attr.Value.Int64())
			}
			break
		}
	}

	if msg == "" {
		return
	}

	_, _ = fmt.Fprint(w, prefix)
	_, _ = fmt.Fprint(w, " ")
	_, _ = fmt.Fprint(w, msg)

	for _, attr := range l.attrs {
		if !strings.HasPrefix(attr.Key, "$") {
			switch attr.Value.Kind() {
			case slog.KindString:
				_, _ = fmt.Fprintf(w, color.WhiteString(" %s=%q", attr.Key, attr.Value))
			default:
				logValue := attr.Value.Any()
				if valuer, ok := logValue.(slog.LogValuer); ok {
					logValue = valuer.LogValue().Any()
				}

				switch x := logValue.(type) {
				case []byte:
					_, _ = fmt.Fprint(w, color.WhiteString(" %s=", attr.Key))
					s := bufio.NewScanner(bytes.NewBuffer(x))
					for s.Scan() {
						if line := s.Text(); len(line) > 0 {
							_, _ = fmt.Fprint(w, color.WhiteString("%s\n", line))
						}
					}
				default:
					_, _ = fmt.Fprint(w, color.WhiteString(" %s=%v", attr.Key, x))
				}
			}
		}
	}
	_, _ = fmt.Fprintln(w)
}

func argsToAttrSlice(args []any) []slog.Attr {
	var (
		attr  slog.Attr
		attrs []slog.Attr
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

func argsToAttr(args []any) (slog.Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]
	case slog.Attr:
		return x, args[1:]
	default:
		return slog.Any(badKey, x), args[1:]
	}
}

const badKey = "!BADKEY"
