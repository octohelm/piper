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
	cueflow.RegisterTask(task.Factory, &Some{})
}

// Some task group
type Some struct {
	task.Group

	// do the step one by one
	Steps []StepInterface `json:"steps"`
	// result values of steps
	Condition []client.Any `json:"-" output:"condition"`
}

func (t *Some) Do(ctx context.Context) error {
	parent := t.Parent()

	scope := cueflow.CueValue(parent.Value().LookupPath(cue.ParsePath("steps")))

	list, err := scope.List()
	if err != nil {
		return err
	}

	for idx := 0; list.Next(); idx++ {
		itemPath := list.Value().Path()

		if err := parent.Scope().RunTasks(ctx, cueflow.WithPrefix(itemPath)); err != nil {
			return errors.Wrapf(err, "steps[%d]", idx)
		}

		stepValue := parent.Scope().Value().LookupPath(itemPath)

		ti := &StepInterface{}
		if err := stepValue.Decode(ti); err != nil {
			return err
		}

		t.Condition = append(t.Condition, client.Any{Value: ti.ResultValue()})

		// when first ok should ok
		if ti.Ok {
			return nil
		}
	}

	return nil
}
