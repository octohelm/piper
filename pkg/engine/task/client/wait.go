package client

import (
	"fmt"
	"github.com/octohelm/cuekit/pkg/cueconvert"
	"strings"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
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

func (ret *WaitInterface) UnmarshalCueValue(v cue.Value) error {
	if v.Kind() != cue.StructKind {
		return fmt.Errorf("client.#Wait must be a struct, but got %s", v.Source())
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
				return fmt.Errorf("task continue, cause got %s", v.Source())
			}
		}

		a := &Any{}
		if err := i.Value().Decode(a); err != nil {
			return cueerrors.Wrapf(err, i.Value().Pos(), "invalid result `%s`", prop)
		}

		ret.values[prop] = a.Value
	}

	return nil
}

var _ cueconvert.OutputValuer = &WaitInterface{}

func (ret *WaitInterface) OutputValues() map[string]any {
	values := map[string]any{}
	for k, v := range ret.values {
		values[k] = v
	}
	return values
}
