package file

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Rel{})
}

// Rel to check path exists
type Rel struct {
	task.Task
	// current workdir
	BaseDir wd.WorkDir `json:"baseDir"`
	// src file
	TargetFile File `json:"targetFile"`
	// the converted file
	File File `json:"-" output:"file"`
}

func (t *Rel) Do(ctx context.Context) error {
	return t.BaseDir.Do(ctx, func(ctx context.Context, baseDir pkgwd.WorkDir) (err error) {
		return t.TargetFile.WorkDir.Do(ctx, func(ctx context.Context, targetDir pkgwd.WorkDir) error {
			baseAddr := baseDir.Addr()
			targetAddr := targetDir.Addr()

			if pkgwd.SameFileSystem(baseAddr, targetAddr) {
				rel, err := filepath.Rel(baseAddr.Path, targetAddr.Path)
				if err != nil {
					return err
				}
				return t.File.Sync(ctx, baseDir, filepath.Join(rel, t.TargetFile.Filename))
			}

			return errors.New("not in same filesystem, please use file.#Sync first")
		})
	})
}
