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
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadFromYAML{})
}

// ReadFromYAML read and parse yaml
type ReadFromYAML struct {
	task.Task

	taskwd.CurrentWorkDir

	// filename
	Filename string `json:"filename"`

	With ReadFromYAMLOption `json:"with"`

	// data
	ReadFromYAMLResult `json:"-" output:"result"`
}

type ReadFromYAMLOption struct {
	// read as list
	AsList bool `json:"asList,omitempty"`
}

type ReadFromYAMLResult struct {
	cueflow.Result

	Data client.Any `json:"data"`
}

func (t *ReadFromYAMLResult) ResultValue() any {
	return t
}

func (t *ReadFromYAML) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer func() {
			t.Done(err)
		}()

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_RDONLY, os.ModePerm)
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
