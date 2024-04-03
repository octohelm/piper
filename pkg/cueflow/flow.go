package cueflow

import (
	"cuelang.org/go/cue"
	"github.com/octohelm/piper/pkg/cueflow/internal"
)

type TaskOptionFunc = internal.OptionFunc

func WithPrefix(path cue.Path) TaskOptionFunc {
	return func(c *internal.Controller) {
		c.Prefix = &path
	}
}
