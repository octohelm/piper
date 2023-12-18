package task

import (
	"fmt"
	"sync"

	"github.com/opencontainers/go-digest"
)

type Store[K comparable, V fmt.Stringer] struct {
	m sync.Map
}

func (s *Store[K, V]) Get(k K) (V, bool) {
	if v, ok := s.m.Load(k); ok {
		return v.(V), true
	}
	return *(new(V)), false
}

func (s *Store[K, V]) Set(v V) string {
	dget := digest.FromString(v.String()).String()
	s.m.Store(dget, v)
	return dget
}
