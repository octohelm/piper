package wd

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Local{})
}

// Local
// create a local workdir
type Local struct {
	task.Task

	// related dir on the root of project
	Source string `json:"source" default:"."`

	// the local workdir
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (local *Local) Do(ctx context.Context) error {
	wd, err := task.ClientContext.From(ctx).SourceDir(ctx)
	if err != nil {
		return err
	}

	finalWd, err := pkgwd.With(wd, pkgwd.WithDir(local.Source))
	if err != nil {
		return err
	}

	return local.WorkDir.Sync(ctx, finalWd)
}
