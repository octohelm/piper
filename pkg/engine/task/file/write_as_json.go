package file

import (
	"context"
	"os"
	"path"

	"encoding/json"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WriteAsJSON{})
}

// WriteAsJSON read and parse json
type WriteAsJSON struct {
	task.Task
	// output file
	OutFile File `json:"outFile"`
	// data could convert to json
	Data client.Any `json:"data"`
	// writen file
	File File `json:"-" output:"file"`
}

func (t *WriteAsJSON) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "open file failed")
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")

		if err := enc.Encode(t.Data.Value); err != nil {
			return errors.Wrap(err, "marshal to json failed")
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
