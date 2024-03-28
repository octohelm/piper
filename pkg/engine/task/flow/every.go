package flow

import (
	"context"
	"cuelang.org/go/cue"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Every{})
}

// Every task group
type Every struct {
	task.Group

	// do the step one by one
	Steps []StepInterface `json:"steps"`
	// result values of steps
	Condition []any `json:"-" output:"condition"`
}

func (t *Every) Do(ctx context.Context) error {
	tt := t.T()

	scope := cueflow.CueValue(tt.Value().LookupPath(cue.ParsePath("steps")))

	list, err := scope.List()
	if err != nil {
		return err
	}

	for idx := 0; list.Next(); idx++ {
		itemValue := list.Value()

		if err := tt.Scope().RunTasks(ctx, cueflow.WithPrefix(itemValue.Path())); err != nil {
			return errors.Wrapf(err, "steps[%d]", idx)
		}

		resultValue := tt.Scope().Value().LookupPath(itemValue.Path())

		ti := &StepInterface{}

		if err := resultValue.Decode(ti); err != nil {

			return err
		}

		t.Condition = append(t.Condition, client.Any{Value: ti.ResultValue()})

		if !ti.Ok {
			return errors.Wrapf(err, "steps[%d] failed", idx)
		}
	}

	return nil
}
