package file

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
)

func init() {
	enginetask.Registry.Register(&Write{})
}

// Write file with contents
type Write struct {
	task.Task

	// output file
	OutFile File `json:"outFile"`
	// file contents
	Contents client.StringOrBytes `json:"contents"`
	// writen file
	File File `json:"-" output:"file"`
}

func (t *Write) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%s: open file failed: %w", cwd, err)
		}
		defer f.Close()

		if _, err = f.Write(t.Contents); err != nil {
			return fmt.Errorf("%s: write file failed: %w", cwd, err)
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
