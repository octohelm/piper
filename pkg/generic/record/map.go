package record

import "sync"

type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Load(k K) (V, bool) {
	value, ok := m.m.Load(k)
	if ok {
		return value.(V), ok
	}
	return *new(V), false
}

func (m *Map[K, V]) Store(k K, v V) {
	m.m.Store(k, v)
}

func (m *Map[K, V]) LoadOrStore(k K, v V) (V, bool) {
	value, ok := m.m.LoadOrStore(k, v)
	if ok {
		return value.(V), ok
	}
	return *new(V), false
}

func (m *Map[K, V]) Delete(k K) {
	m.m.Delete(k)
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
