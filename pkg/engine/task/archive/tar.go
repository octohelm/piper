package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/cueflow"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	"github.com/octohelm/unifs/pkg/filesystem"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Tar{})
}

// Tar
// make a tar archive file of specified dir
type Tar struct {
	task.Task
	// specified dir for tar
	SrcDir wd.WorkDir `json:"srcDir"`
	// tar out filename base on the current work dir
	OutFile file.File `json:"outFile"`
	// created tarfile
	File file.File `json:"-" output:"file"`
}

func (t *Tar) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, outDir pkgwd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, outDir, filepath.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		tarFile, err := outDir.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer tarFile.Close()

		return t.SrcDir.Do(ctx, func(ctx context.Context, srcDir pkgwd.WorkDir) error {
			var w io.WriteCloser = tarFile

			if strings.HasSuffix(t.OutFile.Filename, ".gz") {
				w = gzip.NewWriter(w)
				defer func() {
					_ = w.Close()
				}()
			}

			tw := tar.NewWriter(w)
			defer func() {
				_ = tw.Close()
			}()

			err := pkgwd.ListFile(srcDir, ".", func(filename string) error {
				s, err := srcDir.Stat(ctx, filename)
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

				f, err := srcDir.OpenFile(ctx, filename, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return err
				}
				_, err = io.Copy(tw, f)
				return err
			})

			if err != nil {
				return err
			}

			logr.FromContext(ctx).Info(fmt.Sprintf("%s created.", t.File.Filename))

			return t.File.SyncWith(ctx, t.OutFile)
		})
	})
}
