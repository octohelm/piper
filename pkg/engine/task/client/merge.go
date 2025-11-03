package client

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/x/anyjson"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Merge{})
}

// Merge
// read secret value for the secret
type Merge struct {
	task.Task

	Inputs []Any `json:"inputs"`

	Output Any `json:"-" output:"output"`
}

func (e *Merge) Do(ctx context.Context) error {
	var o anyjson.Valuer
	for _, input := range e.Inputs {
		v, err := anyjson.FromValue(input.Value)
		if err != nil {
			return err
		}
		o = anyjson.Merge(o, v)
	}

	if o != nil {
		e.Output.Value = o.Value()
	}

	return nil
}
