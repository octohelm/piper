package file

import (
	"context"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

func init() {
	cueflow.RegisterTask(task.Factory, &UnmarshalTOML{})
}

// UnmarshalTOML
type UnmarshalTOML struct {
	task.Task
	// toml raw
	Contents client.StringOrBytes `json:"contents"`
	// data
	Data client.Any `json:"-" output:"data"`
}

func (t *UnmarshalTOML) Do(ctx context.Context) error {
	err := toml.Unmarshal(t.Contents, &t.Data.Value)
	if err != nil {
		return errors.Wrap(err, "unmarshal to toml failed")
	}
	return nil
}
