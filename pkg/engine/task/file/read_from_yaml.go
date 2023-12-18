package file

import (
	"context"
	"github.com/octohelm/piper/pkg/anyjson"
	"github.com/octohelm/piper/pkg/engine/task/client"
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

	// data
	ReadFromYAMLResult `json:"-" output:"result"`
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

		o := anyjson.Map{}
		if err := yaml.NewDecoder(f).Decode(&o); err != nil {
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
