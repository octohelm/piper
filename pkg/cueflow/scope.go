package cueflow

import (
	"context"

	"github.com/octohelm/cuekit/pkg/mod/module"

	"cuelang.org/go/cue"
)

type Scope interface {
	Value() Value
	Module() *module.Module
	LookupPath(path cue.Path) Value
	FillPath(path cue.Path, value any) error
	Processed(path cue.Path) bool
	LookupResult(path cue.Path) (any, bool)
	RunTasks(ctx context.Context, optFns ...TaskOptionFunc) error
}
