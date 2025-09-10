package container

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Source{})
}

type Source struct {
	task.Task

	// working dir
	Cwd     wd.WorkDir `json:"cwd"`
	Path    string     `json:"path" default:"."`
	Include []string   `json:"include,omitzero"`
	Exclude []string   `json:"exclude,omitzero"`

	Output Fs `json:"-" output:"output"`
}

func (x *Source) Do(ctx context.Context) error {
	w, err := x.Cwd.Get(ctx)
	if err != nil {
		return fmt.Errorf("%T: get cwd failed: %s", x, err)
	}

	if w.Addr().Scheme != "file" {
		return errors.New("only support local dir as container source")
	}

	base, err := pkgwd.RealPath(w)
	if err != nil {
		return fmt.Errorf("%T: only support cwd in local host", x)
	}

	path := filepath.Join(base, x.Path)

	// storeContainerID the meta until some builder need to use.
	// important for multi-builder build
	return x.Output.SyncLazyDirectory(ctx, x, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		return c.Host().Directory(path, dagger.HostDirectoryOpts{
			Include:         x.Include,
			Exclude:         x.Exclude,
			NoGitAutoIgnore: true,
		}), nil
	})
}
