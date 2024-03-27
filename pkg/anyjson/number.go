package anyjson

import (
	"encoding/json"
	"strconv"
)

type Number struct {
	value any
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
		i, err := strconv.ParseInt(string(v.raw), 10, 64)
		if err == nil {
			v.value = i
		} else {
			f, err := strconv.ParseFloat(string(v.raw), 64)
			if err == nil {
				v.value = f
			}
		}
	}
	return v.value
}

func (v *Number) String() string {
	return ToString(v)
}
