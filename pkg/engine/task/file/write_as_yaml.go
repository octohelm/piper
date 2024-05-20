package file

import (
	"context"
	"os"
	"path"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func init() {
	cueflow.RegisterTask(task.Factory, &WriteAsYAML{})
}

// WriteAsYAML read and parse yaml
type WriteAsYAML struct {
	task.Task

	// output file
	OutFile File `json:"outFile"`
	// options
	With WriteAsYAMLOption `json:"with,omitempty"`
	// data could convert to yaml
	Data client.Any `json:"data"`
	// writen file
	File File `json:"-" output:"file"`
}

type WriteAsYAMLOption struct {
	// write as stream
	AsStream bool `json:"asStream,omitempty"`
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

		enc := yaml.NewEncoder(f)

		switch x := t.Data.Value.(type) {
		case []any:
			if t.With.AsStream {
				for _, item := range x {
					//_, err := fmt.Fprintln(f, "---")
					//if err != nil {
					//	return errors.Wrap(err, "marshal to yaml failed")
					//}
					if err := enc.Encode(item); err != nil {
						return errors.Wrap(err, "marshal to yaml failed")
					}
				}
			} else {
				if err := enc.Encode(x); err != nil {
					return errors.Wrap(err, "marshal to yaml failed")
				}
			}
		default:
			if err := enc.Encode(x); err != nil {
				return errors.Wrap(err, "marshal to yaml failed")
			}
		}

		return t.File.SyncWith(ctx, t.OutFile)
	})
}
