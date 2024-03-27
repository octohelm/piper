package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"path/filepath"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Export{})
}

type Export struct {
	task.Task

	Input   Container `json:"input"`
	OutFile file.File `json:"outFile"`

	File file.File `json:"-" output:"file"`
}

func (x *Export) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		w, err := x.OutFile.WorkDir.Get(ctx)
		if err != nil {
			return err
		}

		base, err := pkgwd.RealPath(w)
		if err != nil {
			return errors.Errorf("%T: only support cwd in local host", x)
		}

		ok, err := x.Input.Container(c).Export(ctx, filepath.Join(base, x.OutFile.Filename), dagger.ContainerExportOpts{
			MediaTypes: dagger.Ocimediatypes,
		})
		if err != nil {
			return err
		}

		if ok {
			return x.File.SyncWith(ctx, x.OutFile)
		}

		return errors.New("export failed")
	})
}
