package client

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"

	"github.com/octohelm/piper/pkg/anyjson"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Merge{})
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
		o = anyjson.Merge(o, anyjson.From(input.Value))
	}
	e.Output.Value = o.Value()
	return nil
}
