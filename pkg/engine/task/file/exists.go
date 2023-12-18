package file

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"os"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Exists{})
}

// Exists to check path exists
type Exists struct {
	task.Task

	taskwd.CurrentWorkDir
	// path
	Path string `json:"path"`

	ExistsResult `json:"-" output:"result"`
}

type ExistsResult struct {
	cueflow.Result

	IsDir bool   `json:"isDir,omitempty"`
	Mode  uint32 `json:"mode,omitempty"`
	Size  int64  `json:"size,omitempty"`
}

func (t *Exists) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		info, err := cwd.Stat(ctx, t.Path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				t.Done(err)
				return nil
			}

			return errors.Wrapf(err, "%s: stat failed", cwd)
		}

		t.Done(nil)

		t.IsDir = info.IsDir()
		t.Mode = uint32(info.Mode())
		t.Size = info.Size()

		return nil
	})
}
