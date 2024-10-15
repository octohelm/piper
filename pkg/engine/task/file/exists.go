package file

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Exists{})
}

// Exists to check path exists
type Exists struct {
	task.Task

	// current workdir
	Cwd wd.WorkDir `json:"cwd"`
	// path
	Path string `json:"path"`

	Info Info `json:"-" output:"info"`
}

type Info struct {
	IsDir bool   `json:"isDir,omitempty"`
	Mode  uint32 `json:"mode,omitempty"`
	Size  int64  `json:"size,omitempty"`
}

func (t *Exists) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) (err error) {
		info, err := cwd.Stat(ctx, t.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("stat failed at %s: %w", cwd, err)
		}

		t.Info.IsDir = info.IsDir()
		t.Info.Mode = uint32(info.Mode())
		t.Info.Size = info.Size()

		return nil
	})
}
