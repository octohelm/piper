package client

import (
	"encoding/json"
	"github.com/octohelm/piper/pkg/anyjson"
	"slices"
)

type Any struct {
	Value any
}

func (Any) CueType() []byte {
	return []byte("_")
}

func (v *Any) UnmarshalJSON(data []byte) error {
	o := &anyjson.Array{}

	if err := json.Unmarshal(slices.Concat([]byte("["), data, []byte("]")), o); err != nil {
		return err
	}

	*v = Any{Value: o.Value().(anyjson.List)[0]}
	return nil
}

func (v Any) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}
