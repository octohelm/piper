package rigutil

import (
	"fmt"
	"github.com/opencontainers/go-digest"
	"sync"
)

type Store[K comparable, V fmt.Stringer] struct {
	m sync.Map
}

func (s *Store[K, V]) Get(k K) (V, bool) {
	if v, ok := s.m.Load(k); ok {
		return v.(V), true
	}
	return *(*V)(nil), false
}

func (s *Store[K, V]) Set(v V) string {
	dget := digest.FromString(v.String()).String()
	s.m.Store(dget, v)
	return dget
}
