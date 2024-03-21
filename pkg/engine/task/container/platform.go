package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Platform{})
}

// Platform resolve platform of container
type Platform struct {
	task.Task

	Input  Container `json:"input"`
	Output string    `json:"-" output:"output"`
}

func (e *Platform) Do(ctx context.Context) error {
	return e.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		container := e.Input.Container(c)

		p, err := container.Platform(ctx)
		if err != nil {
			return err
		}
		e.Output = string(p)

		return nil
	})
}
