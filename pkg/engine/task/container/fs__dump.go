package container

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Dump{})
}

type Dump struct {
	task.Task

	Input Fs `json:"input"`

	OutDir wd.WorkDir `json:"outDir"`

	With DumpOption `json:"with,omitempty"`

	Dir wd.WorkDir `json:"-" output:"dir"`
}

type DumpOption struct {
	Empty bool `json:"empty,omitempty"`
}

func (x *Dump) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		dest, err := x.OutDir.Get(ctx)
		if err != nil {
			return err
		}

		realpath, err := pkgwd.RealPath(dest)
		if err != nil {
			return fmt.Errorf("%T: only support cwd in local host", x)
		}

		if x.With.Empty {
			if err := filesystem.WalkDir(ctx, dest, ".", func(path string, d fs.DirEntry, err error) error {
				// skip root
				if path == "." {
					return nil
				}

				if err := dest.RemoveAll(ctx, path); err != nil {
					return fmt.Errorf("remove failed %s: %w", path, err)
				}
				return filesystem.SkipAll
			}); err != nil {
				return fmt.Errorf("empty outDir failed: %w", err)
			}
		}

		d, err := x.Input.Directory(ctx, c)
		if err != nil {
			return err
		}

		response, err := d.Export(ctx, realpath)
		if err != nil {
			return err
		}

		if len(response) > 0 {
			return x.Dir.Sync(ctx, dest)
		}

		return errors.New("dump failed")
	})
}
