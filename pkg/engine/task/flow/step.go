package flow

import (
	"encoding/json"

	"github.com/octohelm/piper/pkg/cueflow"
)

// StepInterface
// support additional fields but must the `result: ResultInterface` as result checking
type StepInterface struct {
	Result ResultInterface `json:"result"`
}

type Result = cueflow.CanSuccess

var _ Result = ResultInterface{}

// ResultInterface
// support additional fields but the `ok: bool`
type ResultInterface struct {
	cueflow.Result
	values map[string]any
}

func (r ResultInterface) MarshalJSON() ([]byte, error) {
	values := map[string]any{}

	for k, v := range r.values {
		values[k] = v
	}

	values["ok"] = r.Ok

	return json.Marshal(values)
}

func (r *ResultInterface) UnmarshalJSON(bytes []byte) error {
	rr := &ResultInterface{
		values: map[string]any{},
	}

	if err := json.Unmarshal(bytes, &rr.values); err != nil {
		return err
	}

	rr.Ok = rr.values["ok"].(bool)
	*r = *rr
	return nil
}
