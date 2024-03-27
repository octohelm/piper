package file

import (
	"context"
	"github.com/octohelm/piper/pkg/anyjson"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadFromYAML{})
}

// ReadFromYAML read and parse yaml
type ReadFromYAML struct {
	task.Task

	// file
	File File `json:"file"`
	// options
	With ReadFromYAMLOption `json:"with"`
	// data
	Data client.Any `json:"-" output:"data"`
}

type ReadFromYAMLOption struct {
	// read as list
	AsList bool `json:"asList,omitempty"`
}

func (t *ReadFromYAML) Do(ctx context.Context) error {
	return t.File.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		f, err := cwd.OpenFile(ctx, t.File.Filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		list := make([]anyjson.Valuer, 0)

		d := yaml.NewDecoder(f)

		for {
			o := anyjson.Map{}

			err := d.Decode(&o)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			// ignore null value
			v := anyjson.Transform(ctx, anyjson.From(o), func(v anyjson.Valuer, keyPath ...any) anyjson.Valuer {
				if _, ok := v.(*anyjson.Null); ok {
					return nil
				}
				return v
			}).(*anyjson.Object)

			list = append(list, v)
		}

		if n := len(list); !t.With.AsList && n == 1 {
			t.Data.Value = list[0].Value()
		} else {
			values := make([]any, n)
			for i := range list {
				values[i] = list[i].Value()
			}
			t.Data.Value = values
		}

		return nil
	})
}
