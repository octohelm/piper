package task

import (
	"context"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/wd"
	"os"
	"path/filepath"
)

func init() {
	core.DefaultFactory.Register(&WriteFile{})
}

type WriteFile struct {
	core.Task

	CWD      core.WD `json:"cwd"`
	Path     string  `json:"path"`
	Contents string  `json:"contents"`
}

func (e *WriteFile) Do(ctx context.Context) error {
	return e.CWD.Do(ctx, func(cwd wd.WorkDir) error {
		if err := cwd.Mkdir(ctx, filepath.Dir(e.Path), os.ModeDir); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, e.Path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = f.Write([]byte(e.Contents)); err != nil {
			return err
		}

		return nil
	})
}
