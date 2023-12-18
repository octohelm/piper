package chunk

import (
	testingx "github.com/octohelm/x/testing"
	"sort"
	"sync"
	"testing"
)

func TestWorker(t *testing.T) {
	w := NewWorker(21 * MiB)

	values := sync.Map{}

	w.Do(func(c Chunk) error {
		values.Store(c, 1)
		return nil
	})

	err := w.Wait()
	testingx.Expect(t, err, testingx.Be[error](nil))

	chunks := make([]Chunk, 0)

	values.Range(func(key, value any) bool {
		chunks = append(chunks, key.(Chunk))
		return true
	})

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Offset < chunks[j].Offset
	})

	testingx.Expect(t, chunks, testingx.Equal([]Chunk{
		{Offset: 0, Size: 10 * MiB},
		{Offset: 10 * MiB, Size: 10 * MiB},
		{Offset: 20 * MiB, Size: 1 * MiB},
	}))
}
