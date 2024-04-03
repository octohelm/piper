package container

import (
	"context"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Copy{})
}

type Copy struct {
	task.Task

	Input Container `json:"input"`

	Contents Fs       `json:"contents"`
	Source   string   `json:"source" default:"/"`
	Include  []string `json:"include,omitempty"`
	Exclude  []string `json:"exclude,omitempty"`

	Dest string `json:"dest" default:"/"`

	Output Container `json:"-" output:"output"`
}

func (x *Copy) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		base := x.Input.Container(c)
		src, err := x.Contents.Directory(ctx, c)
		if err != nil {
			return err
		}
		return x.Output.Sync(ctx, base.WithDirectory(x.Dest, src, dagger.ContainerWithDirectoryOpts{
			Include: x.Include,
			Exclude: x.Exclude,
		}), x.Input.Platform)
	})
}
