package client

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadSecret{})
}

// ReadSecret
// read secret value for the secret
type ReadSecret struct {
	cueflow.TaskImpl

	// secret ref
	Secret Secret `json:"secret"`

	// secret value
	Value string `json:"-" output:"value"`
}

func (e *ReadSecret) Do(ctx context.Context) error {
	s, ok := e.Secret.Value(ctx)
	if ok {
		e.Value = s.Value
		return nil
	}
	return errors.New("secret not found")
}
