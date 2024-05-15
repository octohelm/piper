package container

import (
	"context"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Diff{})
}

type Diff struct {
	task.Task

	Upper Fs `json:"upper"`
	Lower Fs `json:"lower"`

	Output Fs `json:"-" output:"output"`
}

func (x *Diff) Do(ctx context.Context) error {
	return x.Upper.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		upper, err := x.Upper.Directory(ctx, c)
		if err != nil {
			return err
		}
		lower, err := x.Lower.Directory(ctx, c)
		if err != nil {
			return err
		}
		return x.Output.Sync(ctx, lower.Diff(upper))
	})
}
