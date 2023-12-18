package task

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	core.DefaultFactory.Register(&Tar{})
}

type Tar struct {
	core.Task

	CWD  core.WD `json:"cwd"`
	Path string  `json:"path"`

	Contents core.WD `json:"contents"`
}

func (e *Tar) Do(ctx context.Context) error {
	return e.CWD.Do(ctx, func(cwd wd.WorkDir) error {
		if err := filesystem.MkdirAll(ctx, cwd, filepath.Dir(e.Path)); err != nil {
			return err
		}

		tarFile, err := cwd.OpenFile(ctx, e.Path, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer tarFile.Close()

		return e.Contents.Do(ctx, func(contents wd.WorkDir) error {
			var w io.WriteCloser = tarFile

			if strings.HasSuffix(e.Path, ".gz") {
				w = gzip.NewWriter(w)
				defer func() {
					_ = w.Close()
				}()
			}

			tw := tar.NewWriter(w)
			defer func() {
				_ = tw.Close()
			}()

			err := wd.ListFile(contents, ".", func(filename string) error {
				s, err := contents.Stat(ctx, filename)
				if err != nil {
					return err
				}
				if err := tw.WriteHeader(&tar.Header{
					Name:    filename,
					Size:    s.Size(),
					Mode:    int64(s.Mode()),
					ModTime: s.ModTime(),
				}); err != nil {
					return err
				}

				f, err := contents.OpenFile(ctx, filename, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return err
				}
				_, err = io.Copy(tw, f)
				return err
			})
			if err != nil {
				return err
			}
			return nil
		})
	})
}
