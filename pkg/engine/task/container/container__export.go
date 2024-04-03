package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	"github.com/octohelm/piper/pkg/ociutil"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
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

		cc := x.Input.Container(c)

		destTar := filepath.Join(base, x.OutFile.Filename)

		dest := destTar

		if len(x.Annotations) > 0 {
			dest = destTar + ".wip"
		}

		ok, err := cc.Export(ctx, dest, dagger.ContainerExportOpts{
			MediaTypes: dagger.Ocimediatypes,
		})
		if err != nil {
			return err
		}

		if len(x.Annotations) > 0 {
			defer func() {
				_ = os.RemoveAll(dest)
			}()

			p := &ociutil.OciTarPatcher{
				Manifests: []ociutil.ManifestPatcher{
					{
						Annotations: x.Annotations,
					},
				},
			}

			wipFile, err := os.Open(dest)
			if err != nil {
				return err
			}
			defer wipFile.Close()

			dstFile, err := os.OpenFile(destTar, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer wipFile.Close()

			if err := p.PatchTo(dstFile, wipFile); err != nil {
				return err
			}
		}

		if ok {
			return x.File.SyncWith(ctx, x.OutFile)
		}

		return errors.New("export failed")
	})
}
