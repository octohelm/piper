package file

import (
	"context"
	"io"
	"os"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&ReadAsString{})
}

// ReadAsString read file as string
type ReadAsString struct {
	task.Task
	// file
	File File `json:"file"`
	// text contents
	Contents string `json:"-" output:"contents"`
}

func (t *ReadAsString) Do(ctx context.Context) error {
	return t.File.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		f, err := cwd.OpenFile(ctx, t.File.Filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		t.Contents = string(data)
		return nil
	})
}
