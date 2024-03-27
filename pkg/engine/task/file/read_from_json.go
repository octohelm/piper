package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/octohelm/piper/pkg/anyjson"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadFromJSON{})
}

// ReadFromJSON read and parse json
type ReadFromJSON struct {
	task.Task

	// file
	File File `json:"file"`
	// data
	Data client.Any `json:"-" output:"data"`
}

func (t *ReadFromJSON) Do(ctx context.Context) error {
	return t.File.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		f, err := cwd.OpenFile(ctx, t.File.Filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		o := anyjson.Map{}
		if err := json.NewDecoder(f).Decode(&o); err != nil {
			return err
		}

		// ignore null value
		v := anyjson.Transform(ctx, anyjson.From(o), func(v anyjson.Valuer, keyPath ...any) anyjson.Valuer {
			if _, ok := v.(*anyjson.Null); ok {
				return nil
			}
			return v
		}).(*anyjson.Object)

		t.Data.Value = v.Value()
		return nil
	})
}
