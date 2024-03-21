package container

import (
	"context"
	"dagger.io/dagger"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &HTTPFetch{})
}

type HTTPFetch struct {
	task.Task

	Source string `json:"source"`
	Dest   string `json:"dest"`

	Output Fs `json:"-" output:"output"`
}

func (x *HTTPFetch) ResultValue() any {
	return map[string]any{
		"output": x.Output,
	}
}

func (x *HTTPFetch) Do(ctx context.Context) error {
	return x.Output.SyncLazyDirectory(ctx, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		return c.Directory().WithFile(x.Dest, c.HTTP(x.Source)), nil
	})
}
