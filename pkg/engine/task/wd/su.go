package wd

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Su{})
}

// Su
// switch user
type Su struct {
	task.Task

	// current workdir
	Cwd WorkDir `json:"cwd"`

	// switched user
	User string `json:"user"`

	// switched workdir with the switched user
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (e *Su) Do(ctx context.Context) error {
	return e.Cwd.Do(
		ctx,
		func(ctx context.Context, cwd pkgwd.WorkDir) error {
			return e.WorkDir.Sync(ctx, cwd)
		},
		pkgwd.WithUser(e.User),
	)
}
