package container

import (
	"context"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WriteFile{})
}

type WriteFile struct {
	task.Task

	Input       Fs     `json:"input"`
	Path        string `json:"path"`
	Permissions int    `json:"permissions" default:"0o644"`
	Contents    string `json:"contents"`

	Output Fs `json:"-" output:"output"`
}

func (x *WriteFile) Do(ctx context.Context) error {
	return x.Output.SyncLazyDirectory(ctx, x, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		d, err := x.Input.Directory(ctx, c)
		if err != nil {
			return nil, err
		}
		return d.WithNewFile(x.Path, x.Contents, dagger.DirectoryWithNewFileOpts{
			Permissions: x.Permissions,
		}), nil
	})
}
