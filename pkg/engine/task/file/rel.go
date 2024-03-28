package file

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"path/filepath"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Rel{})
}

// Rel to check path exists
type Rel struct {
	task.Task
	// current workdir
	Cwd wd.WorkDir `json:"cwd"`
	// src file
	SrcFile File `json:"srcFile"`
	// the converted file
	File File `json:"-" output:"file"`
}

func (t *Rel) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) (err error) {
		return t.SrcFile.WorkDir.Do(ctx, func(ctx context.Context, srcDir pkgwd.WorkDir) error {
			cwdAddr := cwd.Addr()
			srcFileAddr := srcDir.Addr()

			if pkgwd.SameFileSystem(cwdAddr, srcFileAddr) {
				rel, err := filepath.Rel(cwdAddr.Path, srcFileAddr.Path)
				if err != nil {
					return err
				}
				return t.File.Sync(ctx, cwd, filepath.Join(rel, t.SrcFile.Filename))
			}

			return errors.New("not in same filesystem, please use file.#Sync first")
		})
	})
}
