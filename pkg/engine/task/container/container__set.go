package container

import (
	"context"
	piperdagger "github.com/octohelm/piper/pkg/dagger"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Set{})
}

type Set struct {
	task.Task

	Input  Container   `json:"input"`
	Config ImageConfig `json:"config"`
	Output Container   `json:"-" output:"output"`
}

func (x *Set) ResultValue() any {
	return map[string]any{
		"output": x.Output,
	}
}

func (x *Set) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		return x.Output.Sync(ctx, x.Config.ApplyTo(x.Input.Container(c)))
	})
}
