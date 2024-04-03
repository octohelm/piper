package cueflow

import (
	"fmt"
	"log/slog"

	"cuelang.org/go/cue"

	encodingcue "github.com/octohelm/piper/pkg/encoding/cue"
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
	case cue.Value:
		return slog.StringValue(WrapValue(x).Source())
	case Value:
		return slog.StringValue(x.Source())
	default:
		data, err := encodingcue.Marshal(c.v)
		if err != nil {
			panic(fmt.Errorf("encoding failed: %s, %v", err, c.v))
		}
		return slog.AnyValue(data)
	}
}
