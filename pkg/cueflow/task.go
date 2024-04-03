package cueflow

import (
	"fmt"
	"os"

	"cuelang.org/go/cue"
	contextx "github.com/octohelm/x/context"
)

var TaskPathContext = contextx.New[string](contextx.WithDefaultsFunc(func() string {
	return ""
}))

var TaskPath = cue.ParsePath("$$task.name")

type Task interface {
	Scope() Scope

	Name() string
	Path() cue.Path
	Deps() []cue.Path
	Value() Value
	Decode(inputs any) error
}

type TaskUnmarshaler interface {
	UnmarshalTask(t Task) error
}

func NewTask(scope Scope, path cue.Path, deps ...cue.Path) Task {
	v := CueValue(scope.LookupPath(path))
	name, _ := v.LookupPath(TaskPath).String()

	return &task{
		name:  name,
		path:  path,
		scope: scope,
		deps:  deps,
	}
}

type task struct {
	name  string
	scope Scope
	path  cue.Path
	deps  []cue.Path
}

func (t *task) Deps() []cue.Path {
	return t.deps
}

func (t *task) Path() cue.Path {
	return t.path
}

func (t *task) Scope() Scope {
	return t.scope
}

func (t *task) Decode(input any) error {
	if u, ok := input.(TaskUnmarshaler); ok {
		return u.UnmarshalTask(t)
	}

	if err := t.Value().Decode(input); err != nil {
		_, _ = fmt.Fprint(os.Stdout, t.Value().Source())
		_, _ = fmt.Fprintln(os.Stdout)
		return err
	}

	return nil
}

func (t *task) Name() string {
	return t.name
}

func (t *task) Value() Value {
	// always pick value from root
	return t.scope.LookupPath(t.Path())
}
