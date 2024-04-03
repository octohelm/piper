package container

import (
	"context"
	"slices"

	"cuelang.org/go/cue"
	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Build{})
}

type StepInterface struct {
	Input  Container `json:"input,omitempty"`
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

	stepIter, err := cueflow.IterSteps(cueflow.CueValue(tt.Value()))
	if err != nil {
		return err
	}

	step := &StepInterface{}

	for idx, itemValue := range stepIter {
		if idx > 0 {
			path := cue.MakePath(slices.Concat(
				itemValue.Path().Selectors(),
				cue.ParsePath("input").Selectors(),
			)...)

			if err := tt.Scope().FillPath(path, step.Output); err != nil {
				return err
			}
		}

		if err := tt.Scope().RunTasks(ctx, cueflow.WithPrefix(itemValue.Path())); err != nil {
			return errors.Wrapf(err, "steps[%d]", idx)
		}

		stepValue := tt.Scope().LookupPath(itemValue.Path())

		if err := stepValue.Decode(step); err != nil {
			return errors.Wrapf(err, "steps[%d]: decode result failed", idx)
		}
	}

	x.Output = step.Output

	return nil
}
