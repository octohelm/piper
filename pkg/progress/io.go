package progress

import (
	"context"
	"io"
	"time"
)

type Reader interface {
	io.Reader

	Process(ctx context.Context) <-chan Progress
}

func NewReader(r io.Reader, total int64) Reader {
	pw := NewWriter(total)

	return &processReader{
		Reader: io.TeeReader(r, pw),
		Writer: pw,
	}
}

type processReader struct {
	io.Reader
	Writer
}

type Writer interface {
	io.Writer
	Process(ctx context.Context) <-chan Progress
}

func NewWriter(total int64) Writer {
	return &processingWriter{size: total}
}

type processingWriter struct {
	written int64
	size    int64
}

func (r *processingWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	r.written += int64(n)
	return n, err
}

func (r *processingWriter) Process(ctx context.Context) <-chan Progress {
	percentCh := make(chan Progress)
	t := time.NewTicker(1 * time.Second)

	go func() {
		defer func() {
			t.Stop()
			close(percentCh)
		}()

		if r.size == 0 {
			return
		}

		for range t.C {
			select {
			case <-ctx.Done():
				return
			default:
				percentCh <- Progress{
					Current: r.written,
					Total:   r.size,
				}
				if r.completed() {
					return
				}
			}
		}
	}()

	return percentCh
}

func (r *processingWriter) completed() bool {
	return r.written >= r.size
}
