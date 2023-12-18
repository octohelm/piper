package client

import (
	"context"
	"encoding/json"
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
	cueflow.TaskImpl

	Inputs []Any `json:"inputs"`
	Output Any   `json:"-" output:"output"`
}

func (e *Merge) ResultValue() any {
	return e.Output
}

func (e *Merge) Do(ctx context.Context) error {
	var o anyjson.Valuer
	for _, input := range e.Inputs {
		o = anyjson.Merge(o, anyjson.From(input.Value))
	}
	e.Output.Value = o.Value()
	return nil
}

type Any struct {
	Value any
}

func (Any) CueType() []byte {
	return []byte("_")
}

func (v *Any) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.Value)
}

func (v Any) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}
