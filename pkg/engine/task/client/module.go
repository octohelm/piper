package client

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Module{})
}

type Module struct {
	task.Task

	// root module
	Module string `json:"-" output:"module"`
	// { dep: version }
	Deps map[string]string `json:"-" output:"deps"`
}

func (t *Module) Do(ctx context.Context) error {
	m := enginetask.ModuleContext.From(ctx)

	t.Module = m.Module

	t.Deps = map[string]string{}

	for name, dep := range m.Deps {
		t.Deps[name] = dep.Version
	}

	return nil
}
