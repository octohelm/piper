package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"path/filepath"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Export{})
}

type Export struct {
	task.Task

	Cwd   wd.WorkDir `json:"cwd"`
	Input Container  `json:"input"`
	Dest  string     `json:"dest,omitempty" default:"x.tar"`

	Result cueflow.Result `json:"-" output:"result"`
}

func (x *Export) ResultValue() any {
	return map[string]any{
		"result": x.Result,
	}
}

func (x *Export) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		w, err := x.Cwd.Get(ctx)
		if err != nil {
			return err
		}

		base, err := pkgwd.RealPath(w)
		if err != nil {
			return errors.Errorf("%T: only support cwd in local host", x)
		}

		ok, err := x.Input.Container(c).Export(ctx, filepath.Join(base, x.Dest), dagger.ContainerExportOpts{
			MediaTypes: dagger.Ocimediatypes,
		})
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
