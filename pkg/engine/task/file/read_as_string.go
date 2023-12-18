package file

import (
	"context"
	"io"
	"os"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadAsString{})
}

// ReadAsString read file as string
type ReadAsString struct {
	task.Task
	taskwd.CurrentWorkDir
	// filename
	Filename string `json:"filename"`

	ReadAsStringResult `json:"-" output:"result"`
}

type ReadAsStringResult struct {
	cueflow.Result

	Contents string `json:"contents"`
}

func (t *ReadAsStringResult) ResultValue() any {
	return t
}

func (t *ReadAsString) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer t.Done(err)

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_RDONLY, os.ModePerm)
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
