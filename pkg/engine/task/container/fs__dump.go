package container

import (
	"context"
	"io/fs"

	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
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
			return errors.Errorf("%T: only support cwd in local host", x)
		}

		if x.With.Empty {
			if err := filesystem.WalkDir(ctx, dest, ".", func(path string, d fs.DirEntry, err error) error {
				if path != "." {
					return nil
				}

				if d.IsDir() {
					return filesystem.SkipAll
				}

				return dest.RemoveAll(ctx, path)
			}); err != nil {
				return errors.Wrap(err, "empty outDir failed")
			}
		}

		d, err := x.Input.Directory(ctx, c)
		if err != nil {
			return err
		}

		ok, err := d.Export(ctx, realpath)
		if err != nil {
			return err
		}

		if ok {
			return x.Dir.Sync(ctx, dest)
		}

		return errors.New("dump failed")
	})
}
