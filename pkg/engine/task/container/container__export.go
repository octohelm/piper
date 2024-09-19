package container

import (
	"context"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Export{})
}

type Export struct {
	task.Task

	Input Container `json:"input"`

	// oci annotations
	Annotations map[string]string `json:"annotations,omitempty"`

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

		cc, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		if len(x.Annotations) > 0 {
			for k, v := range x.Annotations {
				cc = cc.WithAnnotation(k, v)
			}
		}

		resp, err := cc.Export(ctx, filepath.Join(base, x.OutFile.Filename), dagger.ContainerExportOpts{
			MediaTypes: dagger.Ocimediatypes,
		})
		if err != nil {
			return err
		}

		if len(resp) > 0 {
			return x.File.SyncWith(ctx, x.OutFile)
		}

		return errors.New("export failed")
	})
}
