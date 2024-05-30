package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/dagger/dagger/telemetry"
	"github.com/dagger/dagger/telemetry/sdklog"
	"github.com/opencontainers/go-digest"
	"github.com/vito/progrock"
	"github.com/vito/progrock/console"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var daggerDebugEnabled = false

var isTTY = sync.OnceValue(func() bool {
	if os.Getenv("TTY") == "0" {
		return false
	}
	return true
})

func init() {
	if os.Getenv("DAGGER_DEBUG") != "" {
		daggerDebugEnabled = true
	}
}

func New() Exporter {
	return &exporter{}
}

type exporter struct {
	spans               sync.Map
	taskRecorderGetters sync.Map

	client progrock.UIClient
	once   sync.Once

	*progrock.Recorder
}

func (ui *exporter) Shutdown(ctx context.Context) error {
	ui.once.Do(func() {
		_ = ui.Recorder.Close()
	})
	return nil
}

func (ui *exporter) Run(ctx context.Context, run func(ctx context.Context) (rerr error)) error {
	if isTTY() {
		tape := progrock.NewTape()
		tape.Focus(true)
		tape.ShowInternal(daggerDebugEnabled)

		return progrock.DefaultUI().Run(ctx, tape, func(ctx context.Context, client progrock.UIClient) error {
			ui.Recorder = progrock.NewRecorder(tape)
			ui.client = client
			return run(progrock.ToContext(ctx, ui.Recorder))
		})
	}

	w := console.NewWriter(os.Stdout, console.ShowInternal(daggerDebugEnabled))
	ui.Recorder = progrock.NewRecorder(w)
	return run(progrock.ToContext(ctx, ui.Recorder))
}

func (ui *exporter) ExportLogs(ctx context.Context, logs []*sdklog.LogData) error {
	for _, l := range logs {
		if err := ui.export(ctx, l); err != nil {
			return err
		}
	}
	return nil
}

func (ui *exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		ui.spans.Store(span.SpanContext().SpanID(), span)

		total := int64(0)
		internal := false

		for _, attr := range span.Attributes() {
			switch attr.Key {
			case telemetry.ProgressTotalAttr:
				total = attr.Value.AsInt64()
			case "net.peer.name":
				internal = true
			case telemetry.UIEncapsulatedAttr, telemetry.UIInternalAttr:
				internal = true
			case telemetry.DagCallAttr, telemetry.DagDigestAttr, telemetry.DagInputsAttr, telemetry.LLBDigestsAttr:
				internal = true
			default:
			}
		}

		for _, internalSpanPrefix := range []string{
			"moby.buildkit.v1.",
			"moby.filesync.v1.",
		} {
			if strings.HasPrefix(span.Name(), internalSpanPrefix) {
				internal = true
			}
		}

		rec := ui.taskRecorderGetterOf(span.SpanContext().SpanID(), span.Name(), total, internal)()
		if rec == nil {
			continue
		}

		ui.complete(rec, span)
	}

	return nil
}

func (ui *exporter) complete(rec *progrock.TaskRecorder, span sdktrace.ReadOnlySpan) {
	if endTime := span.EndTime(); !endTime.IsZero() {
		rec.Complete()

		if client := ui.client; client != nil {
			client.SetStatusInfo(progrock.StatusInfo{
				Name:  span.Name(),
				Value: fmt.Sprintf("%s", endTime.Sub(span.StartTime())),
			})
		}
	}
}

func (ui *exporter) export(ctx context.Context, logData *sdklog.LogData) error {
	name := ""
	total := int64(0)
	current := int64(0)

	for attr := range logData.WalkAttributes {
		switch attr.Key {
		case LogAttrSpanName:
			name = attr.Value.AsString()
		case telemetry.ProgressTotalAttr:
			total = attr.Value.AsInt64()
		case telemetry.ProgressCurrentAttr:
			current = attr.Value.AsInt64()
		}
	}

	rec := ui.taskRecorderGetterOf(logData.SpanID, name, total, false)()
	if rec == nil {
		return nil
	}

	if current > 0 {
		if total > 0 {
			rec.Progress(current, total)
			return nil
		}

		rec.Current(current)
		return nil
	}

	s := bufio.NewScanner(bytes.NewBufferString(logData.Body().AsString()))
	for s.Scan() {
		if line := s.Text(); len(line) > 0 {
			if _, err := io.WriteString(rec.Stderr(), line+"\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

type TaskRecorderGetter = func() *progrock.TaskRecorder

func (ui *exporter) taskRecorderGetterOf(spanID trace.SpanID, name string, total int64, internal bool) TaskRecorderGetter {
	if v, ok := ui.spans.Load(spanID); ok {
		name = v.(sdktrace.ReadOnlySpan).Name()
	}

	v, _ := ui.taskRecorderGetters.LoadOrStore(spanID, sync.OnceValue(func() (r *progrock.TaskRecorder) {
		if internal {
			return nil
		}

		defer func() {
			r.Start()
		}()

		dgst := digest.Digest(spanID.String())
		if total > 0 {
			return ui.Recorder.Vertex(dgst, name, progrock.Focused()).ProgressTask(total, name)
		}
		return ui.Recorder.Vertex(dgst, name, progrock.Focused()).Task(name)
	}))

	return v.(TaskRecorderGetter)
}
