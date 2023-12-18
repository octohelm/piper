package task

import (
	contextx "github.com/octohelm/x/context"
)

var SecretStore = &Store[string, Secret]{}

var SecretContext = contextx.New[*Store[string, Secret]](
	contextx.WithDefaultsFunc[*Store[string, Secret]](func() *Store[string, Secret] {
		return SecretStore
	}),
)

type Secret struct {
	Key   string
	Value string
}

func (s Secret) String() string {
	return s.Key
}
