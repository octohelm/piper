package container

import (
	"context"
	"time"

	"dagger.io/dagger"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Exec{})
}

type Exec struct {
	task.Task

	Input Container `json:"input"`

	Args       []string                         `json:"args"`
	Mounts     map[string]Mount                 `json:"mounts,omitempty"`
	Env        map[string]client.SecretOrString `json:"env,omitempty"`
	Workdir    string                           `json:"workdir,omitempty" default:"/"`
	Entrypoint []string                         `json:"entrypoint,omitempty"`
	User       string                           `json:"user,omitempty" default:"root:root"`
	Always     bool                             `json:"always,omitempty"`

	Output Container `json:"-" output:"output"`
}

func (e *Exec) Do(ctx context.Context) error {
	return e.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		container := e.Input.Container(c)

		if len(e.Entrypoint) > 0 {
			container = container.WithEntrypoint(e.Entrypoint)
		}

		for n := range e.Mounts {
			mounted, err := e.Mounts[n].MountTo(ctx, c, container)
			if err != nil {
				return err
			}
			container = mounted
		}

		for k := range e.Env {
			if envVar := e.Env[k]; envVar.Secret != nil {
				if s, ok := Secret(ctx, c, envVar.Secret); ok {
					container = container.WithSecretVariable(k, s)
				}
			} else {
				container = container.WithEnvVariable(k, envVar.Value)
			}
		}

		if workdir := e.Workdir; workdir != "" {
			container = container.WithWorkdir(workdir)
		}

		if user := e.User; user != "" {
			container = container.WithUser(user)
		}

		if e.Always {
			// disable cache
			container = container.WithEnvVariable("__RUN_STARTED_AT", time.Now().String())
		}

		container = container.WithExec(e.Args)

		return e.Output.Sync(ctx, container, e.Input.Platform)
	})
}
