package client

import (
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	cueformat "cuelang.org/go/cue/format"
	"github.com/octohelm/cuekit/pkg/cueconvert"
	"github.com/octohelm/cuekit/pkg/cueflow"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&WaitInterface{})
}

// WaitInterface for wait task ready
type WaitInterface struct {
	task.Checkpoint

	// as assertion, one $ok is false
	// all task should break
	Ok bool `json:"$ok" default:"true"`

	values map[string]any
}

var _ cueflow.CueValueUnmarshaler = &WaitInterface{}

func (w *WaitInterface) UnmarshalCueValue(v cue.Value) error {
	if kind := v.Kind(); kind != cue.StructKind {
		raw, _ := cueformat.Node(v.Syntax(
			cue.Concrete(false), // allow incomplete values
			cue.DisallowCycles(true),
			cue.Docs(true),
			cue.All(),
		))

		return fmt.Errorf("client.#Wait must be a struct, but got: \n\n %s \n\n %s", raw, v.Err())
	}

	i, err := v.Fields(cue.All())
	if err != nil {
		return err
	}

	w.values = map[string]any{}

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
				data, _ := cueformat.Node(v.Source())
				return fmt.Errorf("task break, cause got %s", string(data))
			}
		}

		a := &Any{}
		if err := i.Value().Decode(a); err != nil {
			return cueerrors.Wrapf(err, i.Value().Pos(), "invalid result `%s`", prop)
		}

		w.values[prop] = a.Value
	}

	return nil
}

var _ cueconvert.OutputValuer = &WaitInterface{}

func (w *WaitInterface) OutputValues() map[string]any {
	values := map[string]any{}
	for k, v := range w.values {
		values[k] = v
	}
	return values
}
