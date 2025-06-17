package client

import (
	"context"
	"errors"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&ReadSecret{})
}

// ReadSecret
// read secret value for the secret
type ReadSecret struct {
	task.Task

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
