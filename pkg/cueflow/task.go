package cueflow

import (
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
	contextx "github.com/octohelm/x/context"
)

var TaskPathContext = contextx.New[string](contextx.WithDefaultsFunc(func() string {
	return ""
}))

var TaskPath = cue.ParsePath("$$task.name")

type Task interface {
	Name() string
	Path() cue.Path
	Deps() []cue.Path
	Scope() Scope
	Value() Value
	Decode(inputs any) error
	Fill(values map[string]any) error
}

type TaskUnmarshaler interface {
	UnmarshalTask(t Task) error
}

func WrapTask(t *flow.Task, scope Scope) Task {
	name, _ := t.Value().LookupPath(TaskPath).String()

	return &task{
		name:  name,
		scope: scope,
		task:  t,
	}
}

type task struct {
	name  string
	scope Scope
	task  *flow.Task
}

func (t *task) Deps() (paths []cue.Path) {
	deps := t.task.Dependencies()
	paths = make([]cue.Path, len(deps))
	for i := range deps {
		paths[i] = deps[i].Path()
	}
	return nil
}

func (t *task) Path() cue.Path {
	return t.task.Path()
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
	return t.scope.LookupPath(t.task.Path())
}

func (t *task) Fill(values map[string]any) error {
	return t.task.Fill(values)
}
