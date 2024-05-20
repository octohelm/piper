package kubepkg

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-courier/logr"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/octohelm/piper/pkg/otel"
)

func newProcessPool(action string) *processPool {
	return &processPool{
		action:  action,
		workers: make(chan *processWorker),
	}
}

type processPool struct {
	action  string
	workers chan *processWorker
}

func (p *processPool) Close() error {
	close(p.workers)
	return nil
}

func (p *processPool) Progress(ref name.Reference) chan<- v1.Update {
	ch := make(chan v1.Update)
	p.workers <- &processWorker{
		action: p.action,
		ref:    ref,
		ch:     ch,
	}
	return ch
}

func (p *processPool) Wait(ctx context.Context) {
	for w := range p.workers {
		go func() {
			w.Wait(ctx)
		}()
	}
}

type processWorker struct {
	action string
	ref    name.Reference
	ch     chan v1.Update

	lastUpdate atomic.Pointer[v1.Update]
	lastLogAt  atomic.Pointer[time.Time]

	log     logr.Logger
	onceLog sync.Once
}

func (w *processWorker) Wait(ctx context.Context) {
	for u := range w.ch {
		w.lastUpdate.Store(&u)
		w.Log(ctx)
	}
}

func (w *processWorker) Log(ctx context.Context) {
	if u := w.lastUpdate.Load(); u != nil {
		now := time.Now()

		if lastLogAt := w.lastLogAt.Load(); lastLogAt == nil || (now).Sub(*lastLogAt) >= time.Second {
			w.Logger(ctx, u).
				WithValues(slog.Int64(otel.LogAttrProgressTotal, u.Total)).
				WithValues(slog.Int64(otel.LogAttrProgressCurrent, u.Total)).
				Info(w.action)

			w.lastLogAt.Store(&now)
		}

		if u.Total == u.Complete {
			w.Logger(ctx, u).
				WithValues(slog.Int64(otel.LogAttrProgressTotal, u.Total)).
				WithValues(slog.Int64(otel.LogAttrProgressCurrent, u.Total)).
				Info(w.action)
			w.Logger(ctx, u).End()
		}

	}
}

func (w *processWorker) Logger(ctx context.Context, u *v1.Update) logr.Logger {
	w.onceLog.Do(func() {
		_, l := logr.FromContext(ctx).Start(
			ctx,
			w.ref.String(),
			slog.Int64(otel.LogAttrProgressTotal, u.Total),
		)
		w.log = l
	})
	return w.log
}

func newProcessReader(r io.ReadCloser, total int64, update chan<- v1.Update) io.ReadCloser {
	w := &processWriter{
		total:  total,
		update: update,
	}

	return &readCloser{
		Reader: io.TeeReader(r, w),
		close: func() error {
			close(update)
			return r.Close()
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (r *readCloser) Close() error { return r.close() }

type processWriter struct {
	update   chan<- v1.Update
	complete int64
	total    int64
}

func (w *processWriter) Write(p []byte) (int, error) {
	n := len(p)

	w.complete += int64(n)
	w.update <- v1.Update{
		Complete: w.complete,
		Total:    w.total,
	}

	return n, nil
}
