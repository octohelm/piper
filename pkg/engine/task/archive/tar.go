package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/cueflow"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Tar{})
}

// Tar
// make a tar archive file of specified dir
type Tar struct {
	task.Task

	taskwd.CurrentWorkDir

	// specified dir for tar
	Dir taskwd.WorkDir `json:"dir"`

	// tar out filename base on the current work dir
	OutFile string `json:"outFile"`

	// output tar file when created
	// just group cwd and filename
	file.WrittenFileResult `json:"-" output:"result"`
}

func (t *Tar) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, cwd, filepath.Dir(t.OutFile)); err != nil {
			return err
		}

		tarFile, err := cwd.OpenFile(ctx, t.OutFile, os.O_TRUNC|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer tarFile.Close()
		defer func() {
			if err == nil {
				t.File = file.File{
					Wd:       t.Cwd,
					Filename: t.OutFile,
				}

				logr.FromContext(ctx).Info(fmt.Sprintf("%s created.", t.File.Filename))
			}
		}()

		return t.Dir.Do(ctx, func(ctx context.Context, contents wd.WorkDir) error {
			var w io.WriteCloser = tarFile

			if strings.HasSuffix(t.OutFile, ".gz") {
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
