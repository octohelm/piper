package cue

import (
	"cuelang.org/go/cue/cuecontext"
	cueformat "cuelang.org/go/cue/format"
	"cuelang.org/go/encoding/gocode/gocodec"
)

func Marshal(v any) ([]byte, error) {
	codec := gocodec.New(cuecontext.New(), nil)
	val, err := codec.Decode(v)
	if err != nil {
		return nil, err
	}
	return cueformat.Node(val.Syntax(), cueformat.Simplify())
}
