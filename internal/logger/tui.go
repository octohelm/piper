package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dagger/dagger/dagql/idtui"
	"github.com/dagger/dagger/telemetry/sdklog"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func New() *TUI {
	tui := idtui.New()

	return &TUI{
		Frontend: tui,
		out:      colorable.NewColorableStdout(),
		ticker:   time.NewTicker(50 * time.Millisecond),
		done:     make(chan struct{}),
	}
}

type TUI struct {
	*idtui.Frontend

	spans sync.Map
	logs  sync.Map
	idx   int64

	out   io.Writer
	outMu sync.RWMutex

	ticker   *time.Ticker
	done     chan struct{}
	doneOnce sync.Once
	wg       sync.WaitGroup
}

func (t *TUI) doWriter(do func(w io.Writer)) {
	t.outMu.Lock()
	defer t.outMu.Unlock()

	do(t.out)
}

func (t *TUI) Shutdown(ctx context.Context) error {
	t.doneOnce.Do(func() {
		t.ticker.Stop()

		close(t.done)
	})

	t.wg.Wait()

	return t.Frontend.Shutdown(ctx)
}

func (t *TUI) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		t.spans.Store(span.SpanContext().SpanID(), span.Name())
	}
	return t.Frontend.ExportSpans(ctx, spans)
}

func (t *TUI) print(logData *sdklog.LogData) {
	if name, ok := t.spans.Load(logData.SpanID); ok {
		get, _ := t.logs.LoadOrStore(logData.SpanID, sync.OnceValue(func() LogPrinter {
			w := &prefixWriter{
				num:   atomic.AddInt64(&t.idx, 1),
				scope: name.(string),
				buf:   bytes.NewBuffer(nil),
			}

			t.wg.Add(1)

			go func() {
				<-t.done
				t.doWriter(func(out io.Writer) {
					w.emit(out, true)
				})
				t.wg.Done()
			}()

			go func() {
				for range t.ticker.C {
					t.doWriter(func(out io.Writer) {
						w.emit(out, false)
					})
				}
			}()

			return w
		}))

		get.(func() LogPrinter)().PrintLog(logData)
	}
}

func (t *TUI) ExportLogs(ctx context.Context, logs []*sdklog.LogData) error {
	if t.Plain {
		for _, log := range logs {
			t.print(log)
		}
		return nil
	}
	return t.Frontend.ExportLogs(ctx, logs)
}

type prefixWriter struct {
	scope string
	num   int64

	mux            sync.RWMutex
	lineCount      int
	lastCommitedAt time.Time
	buf            *bytes.Buffer
}

func (w *prefixWriter) PrintLog(data *sdklog.LogData) {
	msg := data.Body().AsString()

	if msg == "" {
		return
	}

	s := bufio.NewScanner(bytes.NewBufferString(msg))
	for s.Scan() {
		if line := s.Text(); len(line) > 0 {
			w.collect(line)
		}
	}
}

func (w *prefixWriter) collect(line string) {
	w.mux.Lock()
	defer w.mux.Unlock()

	_, _ = fmt.Fprintf(w.buf, "%s %s\n", color.MagentaString("%d:", w.num), line)
	w.lineCount++
}

func (w *prefixWriter) emit(out io.Writer, final bool) {
	w.mux.Lock()
	defer w.mux.Unlock()

	if w.buf.Len() == 0 {
		return
	}

	if final || w.lineCount > 10 || time.Since(w.lastCommitedAt) > time.Second {
		prefix := color.MagentaString("%d:", w.num)

		_, _ = fmt.Fprintf(out, "%s %s %s\n", prefix, color.CyanString("in"), w.scope)
		_, _ = fmt.Fprintf(out, w.buf.String())
		_, _ = fmt.Fprintln(out)

		w.buf.Reset()
		w.lineCount = 0
		w.lastCommitedAt = time.Now()
	}
}

type LogPrinter interface {
	PrintLog(data *sdklog.LogData)
}
