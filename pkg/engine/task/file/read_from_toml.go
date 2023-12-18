package file

import (
	"context"
	"github.com/octohelm/piper/pkg/anyjson"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/pelletier/go-toml/v2"
	"io"
	"os"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadFromTOML{})
}

// ReadFromTOML read and parse yaml
type ReadFromTOML struct {
	task.Task

	taskwd.CurrentWorkDir

	// filename
	Filename string `json:"filename"`

	With ReadFromTOMLOption `json:"with"`

	// data
	ReadFromTOMLResult `json:"-" output:"result"`
}

type ReadFromTOMLOption struct {
	// read as list
	AsList bool `json:"asList,omitempty"`
}

type ReadFromTOMLResult struct {
	cueflow.Result

	Data client.Any `json:"data"`
}

func (t *ReadFromTOMLResult) ResultValue() any {
	return t
}

func (t *ReadFromTOML) Do(ctx context.Context) error {
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

		d := toml.NewDecoder(f)

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
