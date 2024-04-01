package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Sub{})
}

type Sub struct {
	task.Task

	// source fs
	Input   Fs       `json:"input,omitempty"`
	Source  string   `json:"source" default:"/"`
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
	Dest    string   `json:"dest" default:"/"`
	Output  Fs       `json:"-" output:"output"`
}

func (x *Sub) Do(ctx context.Context) error {
	return x.Output.SyncLazyDirectory(ctx, x, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		base, err := x.Input.Directory(ctx, c)
		if err != nil {
			return nil, err
		}

		return c.Directory().WithDirectory(x.Dest, base.Directory(x.Source), dagger.DirectoryWithDirectoryOpts{
			Include: x.Include,
			Exclude: x.Exclude,
		}), nil
	})
}
