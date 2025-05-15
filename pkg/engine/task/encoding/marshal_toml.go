package file

import (
	"context"
	"fmt"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/pelletier/go-toml/v2"
)

func init() {
	enginetask.Registry.Register(&MarshalTOML{})
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
		return fmt.Errorf("marshal to toml failed: %w", err)
	}
	t.Contents = string(data)
	return nil
}
