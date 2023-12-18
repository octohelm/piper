package task

import (
	"context"
	"github.com/k0sproject/rig"
	"github.com/octohelm/piper/pkg/engine/plan"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/engine/rigutil"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	core.DefaultFactory.Register(&RootfsFromLocal{})
}

type RootfsFromLocal struct {
	core.SetupTask
	Dir     string  `json:"dir" default:"."`
	WorkDir core.WD `json:"-" piper:"generated,name=wd"`
}

func (c *RootfsFromLocal) Do(ctx context.Context) error {
	planRoot := wd.Dir(plan.ContextContext.From(ctx).PlanRoot())

	cwd, err := wd.Wrap(
		&rig.Connection{
			Localhost: &rig.Localhost{
				Enabled: true,
			},
		},
		wd.WithDir(planRoot.String()),
	)
	if err != nil {
		return err
	}

	id := rigutil.WorkDirContext.From(ctx).Set(cwd)
	c.WorkDir = *core.WDOfID(id)

	return nil
}
