package file

import (
	"context"
	"fmt"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Ensure{})
}

// Ensure to check path exists
type Ensure struct {
	task.Task
	// current workdir
	Cwd wd.WorkDir `json:"cwd"`
	// path to file
	Path string `json:"path"`
	// the existed file
	File File `json:"-" output:"file"`
}

func (t *Ensure) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) (err error) {
		if _, err = cwd.Stat(ctx, t.Path); err != nil {
			return fmt.Errorf("stat failed: %w", err)
		}
		return t.File.Sync(ctx, cwd, t.Path)
	})
}
