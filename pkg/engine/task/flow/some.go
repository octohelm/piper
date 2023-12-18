package flow

import (
	"context"
	"strings"

	"cuelang.org/go/cue"
	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Some{})
}

// Some task group
type Some struct {
	task.Group

	// do the step one by one
	Steps []StepInterface `json:"steps"`

	// util some step ok, and remain steps will not execute.
	Result ResultInterface `json:"-" output:"result"`
}

func (t *Some) ResultValue() any {
	return t.Result
}

func (t *Some) Do(ctx context.Context) error {
	p := t.Parent()

	scope := cueflow.CueValue(p.Value().LookupPath(cue.ParsePath("steps")))

	list, err := scope.List()
	if err != nil {
		return err
	}

	for idx := 0; list.Next(); idx++ {
		valuePath := list.Value().Path()

		if err := cueflow.RunTasks(ctx, p.Scope(),
			cueflow.WithShouldRunFunc(func(value cue.Value) bool {
				return value.LookupPath(cueflow.TaskPath).Exists() && strings.HasPrefix(value.Path().String(), valuePath.String())
			}),
			cueflow.WithPrefix(t.Parent().Path()),
		); err != nil {
			return errors.Wrapf(err, "steps[%d]", idx)
		}

		resultValue := p.Scope().Value().LookupPath(valuePath)

		ti := &StepInterface{}
		if err := resultValue.Decode(ti); err != nil {
			return err
		}

		if ti.Result.Success() {
			t.Result = ti.Result
			return nil
		}
	}

	return nil
}
