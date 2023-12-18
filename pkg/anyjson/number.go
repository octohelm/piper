package anyjson

import (
	"encoding/json"
	"strconv"
)

type Number struct {
	value *float64
	raw   []byte
}

func (v *Number) MarshalJSON() ([]byte, error) {
	if v.raw == nil && v.value != nil {
		v.raw, _ = json.Marshal(v.value)
	}
	return v.raw, nil
}

func (v *Number) Value() any {
	if v.value == nil {
		num, _ := strconv.ParseFloat(string(v.raw), 64)
		v.value = &num
	}
	return *v.value
}

func (v *Number) String() string {
	return ToString(v)
}
