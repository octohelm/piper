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
	cueflow.RegisterTask(task.Factory, &Every{})
}

// Every task group
type Every struct {
	task.Group

	// do the step one by one
	Steps []StepInterface `json:"steps"`

	EveryResult `json:"-" output:"result"`
}

// EveryResult util every step ok
type EveryResult struct {
	cueflow.Result

	Results []ResultInterface `json:"results"`
}

func (t *EveryResult) ResultValue() any {
	return t
}

func (t *Every) Do(ctx context.Context) error {
	p := t.Parent()

	scope := cueflow.CueValue(p.Value().LookupPath(cue.ParsePath("steps")))

	list, err := scope.List()
	if err != nil {
		return err
	}

	for idx := 0; list.Next(); idx++ {
		itemValue := list.Value()

		if err := cueflow.RunTasks(ctx, p.Scope(),
			cueflow.WithShouldRunFunc(func(value cue.Value) bool {
				return value.LookupPath(cueflow.TaskPath).Exists() && strings.HasPrefix(value.Path().String(), itemValue.Path().String())
			}),
			cueflow.WithPrefix(t.Parent().Path()),
		); err != nil {
			return errors.Wrapf(err, "steps[%d]", idx)
		}

		resultValue := p.Scope().Value().LookupPath(itemValue.Path())

		ti := &StepInterface{}

		if err := resultValue.Decode(ti); err != nil {

			return err
		}

		t.Results = append(t.Results, ti.Result)

		if !ti.Result.Success() {
			t.Result.Done(errors.New(ti.Result.Reason))
			return nil
		}
	}

	t.Result.Done(nil)

	return nil
}
