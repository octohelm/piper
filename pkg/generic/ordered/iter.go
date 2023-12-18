package ordered

import (
	"context"
	"sort"
)

func IterMap[E any](ctx context.Context, m map[string]E) <-chan *Field[E] {
	ch := make(chan *Field[E])

	if len(m) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer close(ch)

		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			select {
			case <-ctx.Done():
				return
			case ch <- &Field[E]{Key: key, Value: m[key]}:
			}
		}
	}()

	return ch
}

type Field[V any] struct {
	Key   string
	Value V
}
