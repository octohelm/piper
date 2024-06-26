package cueflow

import (
	"cuelang.org/go/cue"
	cueformat "cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
)

type Value interface {
	Path() cue.Path
	Pos() token.Pos

	Exists() bool
	LookupPath(p cue.Path) Value
	FillPath(p cue.Path, v any) Value

	Decode(target any) error
	Source(opts ...cue.Option) string
}

func CueValue(v Value) cue.Value {
	if w, ok := v.(CueValueWrapper); ok {
		return w.CueValue()
	}
	return cue.Value{}
}

type CueValueWrapper interface {
	CueValue() cue.Value
}

func WrapValue(cueValue cue.Value) Value {
	return &value{cueValue: cueValue}
}

type value struct {
	cueValue cue.Value
}

func (val *value) CueValue() cue.Value {
	return val.cueValue
}

func (val *value) Path() cue.Path {
	return val.cueValue.Path()
}

func (val *value) Pos() token.Pos {
	return val.cueValue.Pos()
}

func (val *value) Decode(target any) error {
	return val.cueValue.Decode(target)
}

func (val *value) Source(opts ...cue.Option) string {
	syn := val.cueValue.Syntax(
		append(opts,
			cue.Concrete(false), // allow incomplete values
			cue.DisallowCycles(true),
			cue.Docs(true),
			cue.All(),
		)...,
	)
	data, _ := cueformat.Node(syn, cueformat.Simplify())
	return string(data)
}

func (val *value) Exists() bool {
	return val.cueValue.Exists()
}

func (val *value) LookupPath(p cue.Path) Value {
	return WrapValue(val.cueValue.LookupPath(p))
}

func (val *value) FillPath(p cue.Path, v any) Value {
	return WrapValue(val.cueValue.FillPath(p, v))
}

func IterSteps(value cue.Value) (func(yield func(idx int, item cue.Value) bool), error) {
	v := value.LookupPath(cue.ParsePath("steps"))
	list, err := v.List()
	if err != nil {
		return nil, err
	}
	return func(yield func(idx int, item cue.Value) bool) {
		for idx := 0; list.Next(); idx++ {
			if !yield(idx, list.Value()) {
				return
			}
		}
	}, err
}
