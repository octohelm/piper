package file

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/unifs/pkg/filesystem"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&WriteAsTOML{})
}

// WriteAsTOML read and parse toml
type WriteAsTOML struct {
	task.Task

	// filename
	OutFile File `json:"outFile"`
	// data could convert to yaml
	Data client.Any `json:"data"`
	// writen file
	File File `json:"-" output:"file"`
}

func (t *WriteAsTOML) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, outDir wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, outDir, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}
		f, err := outDir.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		defer f.Close()

		data, err := toml.Marshal(t.Data.Value)
		if err != nil {
			return fmt.Errorf("marshal to toml failed: %w", err)
		}

		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("write data failed: %w", err)
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
