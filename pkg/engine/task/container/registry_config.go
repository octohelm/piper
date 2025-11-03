package container

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Config{})
}

type Config struct {
	task.SetupTask

	Auths map[string]Auth `json:"auths"`
}

func (v *Config) Do(ctx context.Context) error {
	as := RegistryAuthStoreContext.From(ctx)
	for host, a := range v.Auths {
		as.Store(host, &a)
	}
	return nil
}
