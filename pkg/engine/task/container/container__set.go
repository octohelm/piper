package container

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Set{})
}

type Set struct {
	task.Task

	Input  Container   `json:"input"`
	Config ImageConfig `json:"config"`
	Output Container   `json:"-" output:"output"`
}

func (x *Set) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		cc, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		return x.Output.Sync(ctx, x.Config.ApplyTo(cc), x.Input.Platform)
	})
}
