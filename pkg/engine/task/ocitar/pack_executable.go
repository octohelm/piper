package ocitar

import (
	"context"
	"io"
	"os"

	"github.com/octohelm/crkit/pkg/executable"
	"github.com/octohelm/crkit/pkg/ocitar"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &PackExecutable{})
}

type PackExecutable struct {
	task.Task

	Dest string `json:"dest"`

	// [Platform]: _
	Files map[string]file.File `json:"files"`

	// Annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// OutFile of OciTar
	OutFile file.File `json:"outFile"`

	// File of tar created
	File file.File `json:"-" output:"file"`
}

func (t *PackExecutable) Do(ctx context.Context) error {
	packer := &executable.Packer{}

	layers := make([]executable.LayerWithPlatform, 0, len(t.Files))

	for p, f := range t.Files {

		if err := f.WorkDir.Do(ctx, func(ctx context.Context, wd pkgwd.WorkDir) error {
			l, err := executable.PlatformedBinary(p, func() (io.ReadCloser, error) {
				return wd.OpenFile(ctx, f.Filename, os.O_RDONLY, os.ModePerm)
			})
			if err != nil {
				return err
			}

			layers = append(layers, l)
			return nil
		}); err != nil {
			return err
		}

	}

	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) error {
		f, err := cwd.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "open %s failed", t.OutFile.Filename)
		}
		defer f.Close()

		idx, err := packer.PackAsIndexOfOciTar(
			ctx,
			layers,
			executable.WithAnnotations(t.Annotations),
			executable.WithImageName(t.Dest),
		)
		if err != nil {
			return err
		}

		if err := ocitar.Write(f, idx); err != nil {
			return errors.Errorf("%#v", err)
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
