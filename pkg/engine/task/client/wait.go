package client

import (
	"strings"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WaitInterface{})
}

// WaitInterface for wait task ready
type WaitInterface struct {
	task.Checkpoint

	// as assertion, one $ok is false
	// all task should break
	Ok bool `json:"$ok" default:"true"`

	values map[string]any
}

var _ cueflow.TaskUnmarshaler = &WaitInterface{}

func (ret *WaitInterface) UnmarshalTask(t cueflow.Task) error {
	v := cueflow.CueValue(t.Value())

	if v.Kind() != cue.StructKind {
		return errors.Errorf("client.#Wait must be a struct, but got %s", t.Value().Source())
	}

	i, err := v.Fields(cue.All())
	if err != nil {
		return err
	}

	ret.values = map[string]any{}

	for i.Next() {
		s := i.Selector()

		prop := ""
		if s.Type() == cue.StringLabel {
			prop = s.Unquoted()
		} else {
			prop = s.String()
		}

		// avoid task prop and the ok
		if strings.HasPrefix(prop, "$$") {
			continue
		}

		if prop == "$ok" {
			ok, err := i.Value().Bool()
			if err != nil {
				return err
			}

			if !ok {
				return errors.Errorf("task continue, cause got %s", t.Value().Source())
			}
		}

		a := &Any{}
		if err := cueflow.WrapValue(i.Value()).Decode(a); err != nil {
			return cueerrors.Wrapf(err, i.Value().Pos(), "invalid result `%s`", prop)
		}
		ret.values[prop] = a.Value
	}

	return nil
}

var _ cueflow.OutputValuer = &WaitInterface{}

func (ret *WaitInterface) OutputValues() map[string]any {
	values := map[string]any{}
	for k, v := range ret.values {
		values[k] = v
	}
	return values
}
