package flow

import (
	"encoding/json"
)

// StepInterface
// support additional fields but must the `result: Result` as result checking
type StepInterface struct {
	Ok bool `json:"ok,omitempty"`

	Values map[string]any `json:"-"`
}

func (r StepInterface) ResultValue() map[string]any {
	values := map[string]any{}
	for k, v := range r.Values {
		values[k] = v
	}
	values["ok"] = r.Ok
	return values
}

func (r *StepInterface) UnmarshalJSON(bytes []byte) error {
	ret := &struct {
		Ok bool `json:"ok"`
	}{}
	if err := json.Unmarshal(bytes, ret); err != nil {
		return err
	}

	values := map[string]any{}
	if err := json.Unmarshal(bytes, &values); err != nil {
		return err
	}

	delete(values, "ok")

	*r = StepInterface{
		Ok:     ret.Ok,
		Values: values,
	}

	return nil
}
