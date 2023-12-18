package anyjson

import (
	"bytes"
	"strconv"

	"github.com/octohelm/x/ptr"
)

type Boolean struct {
	raw   []byte
	value *bool
}

func (v *Boolean) MarshalJSON() ([]byte, error) {
	if v.raw == nil && v.value != nil {
		v.raw = []byte(strconv.FormatBool(*v.value))
	}
	return v.raw, nil
}

func (v *Boolean) Value() any {
	if v.value == nil {
		v.value = ptr.Ptr(bytes.Equal(v.raw, []byte("true")))
	}
	return *v.value
}

func (v *Boolean) String() string {
	return ToString(v)
}
