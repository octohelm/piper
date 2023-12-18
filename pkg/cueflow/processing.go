package cueflow

import (
	"context"
	"io"
	"time"
)

type ProcessReader interface {
	io.Reader
	Process(ctx context.Context) <-chan Progress
}

func NewProcessReader(r io.Reader, total int64) ProcessReader {
	pw := NewProcessWriter(total)

	return &processReader{
		Reader:        io.TeeReader(r, pw),
		ProcessWriter: pw,
	}
}

type processReader struct {
	io.Reader
	ProcessWriter
}

type ProcessWriter interface {
	io.Writer
	Process(ctx context.Context) <-chan Progress
}

type Progress struct {
	Current int64
	Total   int64
}

func NewProcessWriter(total int64) ProcessWriter {
	return &processingWriter{size: total}
}

type processingWriter struct {
	written int64
	size    int64
}

func (r *processingWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	r.written += int64(n)
	return
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
