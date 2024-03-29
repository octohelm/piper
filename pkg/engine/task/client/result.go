package client

import (
	"context"
	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"strings"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ResultInterface{})
}

// ResultInterface of client
type ResultInterface struct {
	// to avoid added ok
	task.Task

	values map[string]any
}

func (ResultInterface) CacheDisabled() bool {
	return true
}

var _ cueflow.CacheDisabler = &EnvInterface{}

var _ cueflow.TaskUnmarshaler = &ResultInterface{}

func (ret *ResultInterface) UnmarshalTask(t cueflow.Task) error {
	v := cueflow.CueValue(t.Value())

	i, err := v.Fields(cue.All())
	if err != nil {
		return err
	}

	ret.values = map[string]any{}

	for i.Next() {
		prop := i.Selector().Unquoted()

		// avoid task prop and the ok
		if strings.HasPrefix(prop, "$$") || prop == "ok" {
			continue
		}

		a := &Any{}
		if err := cueflow.WrapValue(i.Value()).Decode(a); err != nil {
			return cueerrors.Wrapf(err, i.Value().Pos(), "invalid result `%s`", prop)
		}
		ret.values[prop] = a.Value
	}

	return nil
}

var _ cueflow.OutputValuer = &ResultInterface{}

func (ret *ResultInterface) OutputValues() map[string]any {
	values := map[string]any{}

	for k, v := range ret.values {
		values[k] = v
	}

	values["ok"] = ret.Success()

	return values
}

func (ret *ResultInterface) Do(ctx context.Context) error {
	// nothing to do
	return nil
}
