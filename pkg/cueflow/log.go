package cueflow

import (
	"encoding/json"
	"log/slog"
)

const (
	LogAttrScope           = "$scope"
	LogAttrName            = "$name"
	LogAttrDep             = "$dep"
	LogAttrProgressTotal   = "$progress.total"
	LogAttrProgressCurrent = "$progress.current"
)

func CueLogValue(v any) slog.LogValuer {
	return &logValue{v: v}
}

type logValue struct {
	v any
}

func (c *logValue) LogValue() slog.Value {
	switch x := c.v.(type) {
	case Value:
		data, err := CueValue(x).MarshalJSON()
		if err != nil {
			return slog.AnyValue(x.Source())
		}
		return slog.AnyValue(data)
	default:
		data, err := json.Marshal(c.v)
		if err != nil {
			panic(err)
		}
		return slog.AnyValue(data)
	}
}
