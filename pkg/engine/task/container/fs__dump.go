package container

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"path/filepath"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Dump{})
}

type Dump struct {
	task.Task

	Cwd    wd.WorkDir     `json:"cwd"`
	Input  Fs             `json:"input"`
	Dest   string         `json:"dest,omitempty" default:"."`
	Result cueflow.Result `json:"-" output:"result"`
}

func (x *Dump) ResultValue() any {
	return map[string]any{
		"result": x.Result,
	}
}

func (x *Dump) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		w, err := x.Cwd.Get(ctx)
		if err != nil {
			return err
		}

		base, err := pkgwd.RealPath(w)
		if err != nil {
			return errors.Errorf("%T: only support cwd in local host", x)
		}

		d, err := x.Input.Directory(ctx, c)
		if err != nil {
			return err
		}

		ok, err := d.Export(ctx, filepath.Join(base, x.Dest))
		if err != nil {
			return err
		}

		if ok {
			x.Result.Done(nil)
			return nil
		}

		x.Result.Done(errors.New("export failed"))
		return nil
	})
}
