package container

import (
	"context"
	"path"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &SourceFile{})
}

type SourceFile struct {
	task.Task

	File   file.File `json:"file"`
	Output Fs        `json:"-" output:"output"`
}

func (x *SourceFile) Do(ctx context.Context) error {
	w, err := x.File.WorkDir.Get(ctx)
	if err != nil {
		return errors.Errorf("%T: get cwd failed: %s", x, err)
	}

	if w.Addr().Scheme != "file" {
		return errors.New("only support local dir as container source")
	}

	base, err := pkgwd.RealPath(w)
	if err != nil {
		return errors.Errorf("%T: only support cwd in local host", x)
	}

	srcDir := filepath.Join(base, path.Dir(x.File.Filename))
	srcFile := path.Base(x.File.Filename)

	// storeContainerID the meta until some builder need to use.
	// important for multi-builder build
	return x.Output.SyncLazyDirectory(ctx, x, func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
		return c.Host().Directory(srcDir, dagger.HostDirectoryOpts{
			Include: []string{
				srcFile,
			},
		}), nil
	})
}
