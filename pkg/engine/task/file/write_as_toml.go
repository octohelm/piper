package file

import (
	"context"
	"os"
	"path"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WriteAsTOML{})
}

// WriteAsYAML read and parse yaml
type WriteAsTOML struct {
	task.Task

	taskwd.CurrentWorkDir
	// filename
	Filename string `json:"filename"`
	// data could convert to yaml
	Data client.Any `json:"data"`

	WrittenFileResult `json:"-" output:"result"`
}

func (t *WriteAsTOML) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer t.Done(err)

		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "open file failed")
		}
		defer f.Close()

		defer func() {
			if err == nil {
				t.WrittenFileResult.Ok = true
				t.WrittenFileResult.File.Cwd = t.Cwd
				t.WrittenFileResult.File.Filename = t.Filename
			}
		}()

		data, err := toml.Marshal(t.Data.Value)
		if err != nil {
			return errors.Wrap(err, "marshal to toml failed")
		}

		if _, err := f.Write(data); err != nil {
			return errors.Wrap(err, "write data failed")
		}

		return
	})
}
