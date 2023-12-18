package client

import (
	"context"

	"github.com/octohelm/piper/pkg/engine/task"
)

func SecretOfID(id string) *Secret {
	s := &Secret{}
	s.Ref.ID = id
	return s
}

type Secret struct {
	Ref struct {
		ID string `json:"id,omitempty"`
	} `json:"$$secret"`
}

func (s *Secret) Value(ctx context.Context) (task.Secret, bool) {
	return task.SecretContext.From(ctx).Get(s.Ref.ID)
}
