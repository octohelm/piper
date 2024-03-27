package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/unifs/pkg/filesystem"
	"io"
	"os"

	"github.com/octohelm/piper/pkg/engine/task"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &UnTar{})
}

// UnTar
// un tar files into specified dest
type UnTar struct {
	task.Task

	taskwd.CurrentWorkDir

	// tar filename base on the current work dest
	Filename string `json:"filename"`
	// tar file content encoding
	ContentEncoding string `json:"contentEncoding,omitempty"`

	// output dest for tar
	OutDir taskwd.WorkDir `json:"outDir"`

	Result cueflow.Result `json:"-" output:"result"`
}

func (t *UnTar) ResultValue() any {
	return t.Result
}

func (t *UnTar) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer func() {
			t.Result.Done(err)
		}()

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		var r io.Reader = f

		if t.ContentEncoding == "gzip" {
			gr, err := gzip.NewReader(f)
			if err != nil {
				return err
			}
			defer gr.Close()
			r = gr
		}

		tarReader := tar.NewReader(r)

		return t.OutDir.Do(ctx, func(ctx context.Context, dest wd.WorkDir) error {
			sync := func(ctx context.Context, hdr *tar.Header) error {
				fi := hdr.FileInfo()

				if fi.IsDir() {
					return filesystem.MkdirAll(ctx, dest, fi.Name())
				}

				f, err := dest.OpenFile(ctx, fi.Name(), os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
				if err != nil {
					return err
				}
				defer f.Close()

				if _, err := io.Copy(f, tarReader); err != nil {
					return err
				}
				return nil
			}

			for {
				hdr, err := tarReader.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				if err := sync(ctx, hdr); err != nil {
					return err
				}

			}

			return nil
		})
	})
}
