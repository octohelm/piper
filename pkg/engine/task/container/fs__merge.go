package container

import (
	"context"
	"errors"

	"dagger.io/dagger"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Merge{})
}

type Merge struct {
	task.Task

	Inputs []Fs `json:"inputs"`
	Output Fs   `json:"-" output:"output"`
}

func (x *Merge) Do(ctx context.Context) error {
	if len(x.Inputs) == 0 {
		return errors.New("empty inputs")
	}

	return x.Inputs[0].Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		d := c.Directory()

		for _, input := range x.Inputs {
			inputDir, err := input.Directory(ctx, c)
			if err != nil {
				return err
			}
			d = d.WithDirectory("/", inputDir)
		}

		return x.Output.Sync(ctx, d)
	})
}
