package file

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WriteAsYAML{})
}

// WriteAsYAML read and parse yaml
type WriteAsYAML struct {
	task.Task

	// output file
	OutFile File `json:"outFile"`
	// data could convert to yaml
	Data client.Any `json:"data"`
	// writen file
	File File `json:"-" output:"file"`
}

func (t *WriteAsYAML) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, outDir wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, outDir, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := outDir.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "open file failed")
		}
		defer f.Close()

		data, err := yaml.Marshal(t.Data.Value)
		if err != nil {
			return errors.Wrap(err, "marshal to yaml failed")
		}

		if _, err := f.Write(data); err != nil {
			return errors.Wrap(err, "write data failed")
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
