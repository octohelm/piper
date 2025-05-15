package wd

import (
	"context"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Sub{})
}

// Sub
// create new work dir base on current work dir
type Sub struct {
	task.Task

	// current workdir
	Cwd WorkDir `json:"cwd"`

	// related path from current workdir
	Path string `json:"path"`

	// new workdir
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (e *Sub) Do(ctx context.Context) error {
	return e.Cwd.Do(
		ctx,
		func(ctx context.Context, cwd wd.WorkDir) error {
			return e.WorkDir.Sync(ctx, cwd)
		},
		wd.WithDir(e.Path),
	)
}
