package file

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
	"os"
	"path"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Write{})
}

// Write file with contents
type Write struct {
	task.Task

	taskwd.CurrentWorkDir

	// filename
	Filename string `json:"filename"`
	// file contents
	Contents client.StringOrBytes `json:"contents"`

	// the written file
	// just group cwd and filename
	WrittenFileResult `json:"-" output:"result"`
}

func (t *Write) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer t.Done(err)

		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "%s: open file failed", cwd)
		}
		defer f.Close()

		if _, err = f.Write(t.Contents); err != nil {
			return errors.Wrapf(err, "%s: write file failed", cwd)
		}

		t.WrittenFileResult.Ok = true
		t.WrittenFileResult.File.Cwd = t.Cwd
		t.WrittenFileResult.File.Filename = t.Filename

		return nil
	})
}

type WrittenFileResult struct {
	cueflow.Result

	File File `json:"file"`
}

func (t *WrittenFileResult) ResultValue() any {
	return t
}
