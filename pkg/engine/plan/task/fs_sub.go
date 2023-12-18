package task

import (
	"context"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	core.DefaultFactory.Register(&Sub{})
}

type Sub struct {
	core.Task
	CWD     core.WD `json:"cwd"`
	Path    string  `json:"path"`
	WorkDir core.WD `json:"-" piper:"generated,name=wd"`
}

func (e *Sub) Do(ctx context.Context) error {
	return e.CWD.Do(ctx, func(cwd wd.WorkDir) error {
		e.WorkDir.SetBy(ctx, cwd)
		return nil
	}, wd.WithDir(e.Path))
}
