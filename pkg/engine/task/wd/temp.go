package wd

import (
	"context"
	"path/filepath"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Temp{})
}

// Temp
// create a tmp workdir
type Temp struct {
	task.Task
	// related dir on the root of project
	ID string `json:"id"`
	// the tmp workdir
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (local *Temp) Do(ctx context.Context) error {
	wd, err := enginetask.ClientContext.From(ctx).SourceDir(ctx)
	if err != nil {
		return err
	}

	finalWd, err := pkgwd.With(wd, pkgwd.WithDir(filepath.Join(".piper", local.ID)))
	if err != nil {
		return err
	}

	return local.WorkDir.Sync(ctx, finalWd)
}
