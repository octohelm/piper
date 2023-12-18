package task

import (
	"context"
	"github.com/octohelm/cuemod/pkg/modutil"
	"github.com/octohelm/piper/pkg/engine/plan"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
)

func init() {
	core.DefaultFactory.Register(&RevInfo{})
}

type RevInfo struct {
	core.SetupTask

	Version string `json:"-" piper:"generated,name=version"`
}

func (e *RevInfo) Do(ctx context.Context) error {
	planRoot := plan.ContextContext.From(ctx).PlanRoot()
	r, err := modutil.RevInfoFromDir(context.Background(), planRoot)
	if err != nil {
		return err
	}
	e.Version = r.Version
	return nil
}
