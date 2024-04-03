package container

import (
	"context"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadFile{})
}

type ReadFile struct {
	task.Task

	Input Fs     `json:"input"`
	Path  string `json:"path"`

	Contents string `json:"-" output:"contents"`
}

func (x *ReadFile) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		dir, err := x.Input.Directory(ctx, c)
		if err != nil {
			return err
		}
		contents, err := dir.File(x.Path).Contents(ctx)
		if err != nil {
			return err
		}
		x.Contents = contents
		return nil

	})
}
