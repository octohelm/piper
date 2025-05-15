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
	"gopkg.in/yaml.v3"
)

func init() {
	enginetask.Registry.Register(&WriteAsYAML{})
}

// WriteAsYAML read and parse yaml
type WriteAsYAML struct {
	task.Task

	// output file
	OutFile File `json:"outFile"`
	// options
	With WriteAsYAMLOption `json:"with,omitzero"`
	// data could convert to yaml
	Data client.Any `json:"data"`
	// writen file
	File File `json:"-" output:"file"`
}

type WriteAsYAMLOption struct {
	// write as stream
	AsStream bool `json:"asStream,omitzero"`
}

func (t *WriteAsYAML) Do(ctx context.Context) error {
	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, outDir wd.WorkDir) (err error) {
		if err := filesystem.MkdirAll(ctx, outDir, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := outDir.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		defer f.Close()

		enc := yaml.NewEncoder(f)

		switch x := t.Data.Value.(type) {
		case []any:
			if t.With.AsStream {
				for _, item := range x {
					if err := enc.Encode(item); err != nil {
						return fmt.Errorf("marshal to yaml failed: %w", err)
					}
				}
			} else {
				if err := enc.Encode(x); err != nil {
					return fmt.Errorf("marshal to yaml failed: %w", err)
				}
			}
		default:
			if err := enc.Encode(x); err != nil {
				return fmt.Errorf("marshal to yaml failed: %w", err)
			}
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
