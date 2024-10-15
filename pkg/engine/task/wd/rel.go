package wd

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Rel{})
}

// Rel to get related path between two dirs
type Rel struct {
	task.Task

	BaseDir   WorkDir `json:"baseDir"`
	TargetDir WorkDir `json:"targetDir"`

	Path string `json:"-" output:"path"`
}

func (t *Rel) Do(ctx context.Context) error {
	return t.BaseDir.Do(ctx, func(ctx context.Context, base pkgwd.WorkDir) (err error) {
		return t.TargetDir.Do(ctx, func(ctx context.Context, target pkgwd.WorkDir) error {
			baseAddr := base.Addr()
			targetAddr := target.Addr()

			if pkgwd.SameFileSystem(baseAddr, targetAddr) {
				rel, err := filepath.Rel(baseAddr.Path, targetAddr.Path)
				if err != nil {
					return err
				}
				t.Path = rel
				return nil
			}

			return errors.New("not in same filesystem, please use file.#Sync first")
		})
	})
}
