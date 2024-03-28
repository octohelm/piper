package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Rm{})
}

type Rm struct {
	task.Task

	Input  Fs                   `json:"input"`
	Path   client.StringOrSlice `json:"path"`
	Output Fs                   `json:"-" output:"output"`
}

func (x *Rm) Do(ctx context.Context) error {
	return x.Output.SyncLazyDirectory(ctx, x, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		dir, err := x.Input.Directory(ctx, c)
		if err != nil {
			return nil, err
		}
		for _, p := range x.Path {
			dir = dir.WithoutDirectory(p)
		}
		return dir, nil
	})
}
