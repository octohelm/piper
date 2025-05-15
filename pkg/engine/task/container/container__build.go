package container

import (
	"context"
	"fmt"
	"slices"

	"cuelang.org/go/cue"
	"github.com/octohelm/cuekit/pkg/cueflow"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/cuekit/pkg/cuepath"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Build{})
}

type StepInterface struct {
	Input  Container `json:"input,omitzero"`
	Output Container `json:"output"`
}

// Build docker build step
type Build struct {
	task.Group

	Steps  []StepInterface `json:"steps"`
	Output Container       `json:"-" output:"output"`
}

func (x *Build) Do(ctx context.Context) error {
	tt := x.T()

	step := &StepInterface{}

	for stepValue, err := range task.Steps(tt.Value()) {
		if err != nil {
			return err
		}

		stepPath := stepValue.Path()
		selectors := stepPath.Selectors()
		idx := selectors[len(selectors)-1].Index()

		if idx > 0 {
			path := cue.MakePath(slices.Concat(
				stepValue.Path().Selectors(),
				cue.ParsePath("input").Selectors(),
			)...)

			if err := tt.Scope().FillPath(path, step.Output); err != nil {
				return err
			}
		}

		if err := cueflow.RunSubTasks(ctx, tt.Scope(), func(p cue.Path) bool {
			return cuepath.Prefix(p, stepPath)
		}); err != nil {
			return err
		}

		if err := tt.Scope().LookupPath(stepValue.Path()).Decode(step); err != nil {
			return fmt.Errorf("steps[%d]: decode result failed: %w", idx, err)
		}
	}

	x.Output = step.Output

	return nil
}
