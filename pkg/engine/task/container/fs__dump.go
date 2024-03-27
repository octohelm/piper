package container

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Dump{})
}

type Dump struct {
	task.Task

	Input Fs `json:"input"`

	OutDir wd.WorkDir `json:"outDir"`

	Dir wd.WorkDir `json:"-" output:"dir"`
}

func (x *Dump) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		w, err := x.OutDir.Get(ctx)
		if err != nil {
			return err
		}

		realpath, err := pkgwd.RealPath(w)
		if err != nil {
			return errors.Errorf("%T: only support cwd in local host", x)
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
			return x.Dir.Sync(ctx, w)
		}

		return errors.New("dump failed")
	})
}
