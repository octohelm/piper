package container

import (
	"context"
	"cuelang.org/go/cue"
	"dagger.io/dagger"
	"fmt"
	"github.com/octohelm/cuekit/pkg/cueflow"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/cuekit/pkg/cuepath"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&RootfsDo{})
}

type RootfsDoStepInterface struct {
	Input  Fs `json:"input,omitzero"`
	Output Fs `json:"output"`
}

// FsDo docker build but with rootfs
type RootfsDo struct {
	task.Group

	Input  Container               `json:"input"`
	Steps  []RootfsDoStepInterface `json:"steps"`
	Output Container               `json:"-" output:"output"`
}

func (x *RootfsDo) Do(ctx context.Context) error {
	tt := x.T()

	path := cuepath.Join(tt.Path(), cue.ParsePath("input"))

	if err := tt.Scope().DecodePath(path, &x.Input); err != nil {
		return err
	}

	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		cc, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		step := &RootfsDoStepInterface{}
		step.Output = x.Input.Rootfs

		for stepValue, err := range task.Steps(tt.Scope().LookupPath(tt.Path())) {
			if err != nil {
				return err
			}

			stepPath := stepValue.Path()
			selectors := stepPath.Selectors()
			idx := selectors[len(selectors)-1].Index()

			p := cuepath.Join(stepValue.Value().Path(), cue.ParsePath("input"))

			if err := tt.Scope().FillPath(p, step.Output); err != nil {
				return err
			}

			if err := cueflow.RunSubTasks(ctx, tt.Scope(), func(p cue.Path) (bool, cue.Path) {
				if cuepath.Prefix(p, stepPath) {
					return true, cuepath.TrimPrefix(p, stepPath)
				}
				return false, cue.MakePath()
			}); err != nil {
				return err
			}

			if err := tt.Scope().DecodePath(stepValue.Path(), step); err != nil {
				return fmt.Errorf("steps[%d]: decode result failed: %w", idx, err)
			}
		}

		finalRootfs, err := step.Output.Directory(ctx, c)
		if err != nil {
			return err
		}

		return x.Output.Sync(ctx, cc.WithRootfs(finalRootfs), x.Input.Platform)
	})
}
