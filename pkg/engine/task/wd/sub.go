package wd

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"

	"github.com/octohelm/piper/pkg/engine/task"

	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Sub{})
}

// Sub
// create new work dir base on current work dir
type Sub struct {
	task.Task

	CurrentWorkDir

	// related path from current work dir
	Dir string `json:"dir"`

	// new work dir
	WorkDir WorkDir `json:"-" output:"wd"`
}

func (e *Sub) Do(ctx context.Context) error {
	return e.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) error {
		e.WorkDir.SetBy(ctx, cwd)
		return nil
	}, wd.WithDir(e.Dir))
}
