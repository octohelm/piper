package container

import (
	"context"
	"slices"
	"time"

	"dagger.io/dagger"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

func init() {
	enginetask.Registry.Register(&Exec{})
}

type Exec struct {
	task.Task

	Input Container `json:"input"`

	Args       []string                         `json:"args"`
	Mounts     map[string]Mount                 `json:"mounts,omitzero"`
	Env        map[string]client.SecretOrString `json:"env,omitzero"`
	Workdir    string                           `json:"workdir,omitzero" default:"/"`
	Entrypoint []string                         `json:"entrypoint,omitzero"`
	User       string                           `json:"user,omitzero" default:"root:root"`
	Always     bool                             `json:"always,omitzero"`

	Output Container `json:"-" output:"output"`
}

func (e *Exec) Do(ctx context.Context) error {
	return e.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		container, err := e.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		execOptions := make([]dagger.ContainerWithExecOpts, 0)

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

		container = container.WithExec(slices.Concat(e.Entrypoint, e.Args), execOptions...)

		if err := e.Output.Sync(ctx, container, e.Input.Platform); err != nil {
			return err
		}

		return nil
	})
}
