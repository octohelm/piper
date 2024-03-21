package container

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Dir{})
}

type Dir struct {
	task.Group

	Input Container `json:"input,omitempty"`
	Path  string    `json:"path" default:"/"`

	Output Fs `json:"-" output:"output"`
}

func (x *Dir) ResultValue() any {
	return map[string]any{
		"output": x.Output,
	}
}

func (x *Dir) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		inputContainer := x.Input.Container(c)

		return x.Output.Sync(ctx, inputContainer.Rootfs().Directory(x.Path))
	})
}
