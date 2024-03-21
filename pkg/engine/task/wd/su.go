package wd

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Su{})
}

// Su
// switch user
type Su struct {
	task.Task

	CurrentWorkDir

	// new user
	User string `json:"user"`

	// new work dir
	WorkDir WorkDir `json:"-" output:"wd"`
}

func (e *Su) Do(ctx context.Context) error {
	return e.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) error {
		return e.WorkDir.Sync(ctx, cwd)
	}, wd.WithUser(e.User))
}
