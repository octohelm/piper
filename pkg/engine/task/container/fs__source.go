package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"path/filepath"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Source{})
}

type Source struct {
	task.Task

	// working dir
	Cwd     wd.WorkDir `json:"cwd"`
	Path    string     `json:"path" default:"."`
	Include []string   `json:"include"`
	Exclude []string   `json:"exclude"`

	Output Fs `json:"-" output:"output"`
}

func (x *Source) ResultValue() any {
	return map[string]any{
		"output": x.Output,
	}
}

func (x *Source) Do(ctx context.Context) error {
	w, err := x.Cwd.Get(ctx)
	if err != nil {
		return errors.Errorf("%T: get cwd failed: %s", x, err)
	}

	base, err := pkgwd.RealPath(w)
	if err != nil {
		return errors.Errorf("%T: only support cwd in local host", x)
	}

	path := filepath.Join(base, x.Path)

	// store the meta until some builder need to use.
	// important for multi-builder build
	return x.Output.SyncLazyDirectory(ctx, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		return c.Host().Directory(path, dagger.HostDirectoryOpts{
			Include: x.Include,
			Exclude: x.Exclude,
		}), nil
	})
}