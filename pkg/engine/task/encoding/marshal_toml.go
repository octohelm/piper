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
	cueflow.RegisterTask(task.Factory, &MarshalTOML{})
}

// MarshalTOML
type MarshalTOML struct {
	task.Task
	// data
	Data client.Any `json:"data"`
	// raw
	Contents string `json:"-" output:"contents"`
}

func (t *MarshalTOML) Do(ctx context.Context) error {
	data, err := toml.Marshal(t.Data.Value)
	if err != nil {
		return errors.Wrap(err, "marshal to toml failed")
	}
	t.Contents = string(data)
	return nil
}
