package chunk

import (
	"sync"
)

type OptionFunc func(w *Worker)

func WithChunkSize(chunkSize FileSize) OptionFunc {
	return func(w *Worker) {
		if chunkSize > 0 {
			w.chunkSize = chunkSize
		}
	}
}

func WithMaxConcurrent(maxConcurrent int) OptionFunc {
	return func(w *Worker) {
		if maxConcurrent > 0 {
			w.maxConcurrent = maxConcurrent
		}
	}
}

func NewWorker(total FileSize, optFns ...OptionFunc) *Worker {
	w := &Worker{
		total:         total,
		chunkSize:     10 * MiB,
		maxConcurrent: 16,

		chunkQueue: make(chan Chunk),
	}

	for _, fn := range optFns {
		fn(w)
	}

	return w
}

type Worker struct {
	total         FileSize
	chunkSize     FileSize
	maxConcurrent int

	once sync.Once

	wg       sync.WaitGroup
	doneOnce sync.Once
	err      error

	chunkQueue chan Chunk
	offset     FileSize
}

func (w *Worker) next() bool {
	offset := w.offset
	if offset == w.total {
		w.doneOnce.Do(func() {
			close(w.chunkQueue)
		})
		return false
	}

	remain := w.total - offset

	if remain < w.chunkSize {
		w.chunkQueue <- Chunk{
			Offset: offset,
			Size:   remain,
		}
		w.offset += remain
	} else {
		w.chunkQueue <- Chunk{
			Offset: offset,
			Size:   w.chunkSize,
		}
		w.offset += w.chunkSize
	}

	return true
}

func (w *Worker) Do(action func(c Chunk) error) {
	w.once.Do(func() {
		for i := 0; i < w.maxConcurrent; i++ {
			w.wg.Add(1)
			go func() {
				defer w.wg.Done()

				for c := range w.chunkQueue {
					if err := action(c); err != nil {
						w.doneOnce.Do(func() {
							close(w.chunkQueue)
							w.err = err
						})
						return
					}
				}
			}()
		}

		go func() {
			for w.next() {

			}
		}()
	})
}

func (w *Worker) Wait() error {
	w.wg.Wait()
	return w.err
}

type Chunk struct {
	Offset FileSize
	Size   FileSize
}
