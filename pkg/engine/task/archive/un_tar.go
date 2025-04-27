package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &UnTar{})
}

// UnTar
// un tar files into specified outDir
type UnTar struct {
	task.Task

	// tar filename base on the current work outDir
	SrcFile file.File `json:"srcFile"`
	// tar file content encoding
	ContentEncoding string `json:"contentEncoding,omitzero"`
	// output outDir for tar
	OutDir wd.WorkDir `json:"outDir"`
	// final dir contains tar files
	Dir wd.WorkDir `json:"-" output:"dir"`
}

func (t *UnTar) Do(ctx context.Context) error {
	return t.SrcFile.WorkDir.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) error {
		f, err := cwd.OpenFile(ctx, t.SrcFile.Filename, os.O_RDONLY, os.ModePerm)
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

		return t.OutDir.Do(ctx, func(ctx context.Context, outDir pkgwd.WorkDir) error {
			sync := func(ctx context.Context, hdr *tar.Header) error {
				fi := hdr.FileInfo()

				if fi.IsDir() {
					return filesystem.MkdirAll(ctx, outDir, fi.Name())
				}

				f, err := outDir.OpenFile(ctx, fi.Name(), os.O_TRUNC|os.O_RDWR|os.O_CREATE, fi.Mode())
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

			return t.Dir.Sync(ctx, outDir)
		})
	})
}
