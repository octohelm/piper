package client

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Module{})
}

type Module struct {
	task.Task
	t cueflow.Task

	// root module
	Module string `json:"-" output:"module"`
	// { dep: version }
	Deps map[string]string `json:"-" output:"deps"`
}

var _ cueflow.TaskUnmarshaler = &Module{}

func (v *Module) UnmarshalTask(t cueflow.Task) error {
	v.t = t
	return nil
}

func (t *Module) Do(ctx context.Context) error {
	m := t.t.Scope().Module()

	t.Module = m.Module
	t.Deps = map[string]string{}
	for name, dep := range m.Deps {
		t.Deps[name] = dep.Version
	}

	return nil
}
