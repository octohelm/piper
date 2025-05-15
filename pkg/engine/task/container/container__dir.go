package container

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Dir{})
}

type Dir struct {
	task.Task

	Input Container `json:"input,omitzero"`
	Path  string    `json:"path" default:"/"`

	Output Fs `json:"-" output:"output"`
}

func (x *Dir) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		inputContainer, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		return x.Output.Sync(ctx, inputContainer.Rootfs().Directory(x.Path))
	})
}
