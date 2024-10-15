package container

import (
	"context"
	"fmt"
	"slices"

	"dagger.io/dagger"

	"cuelang.org/go/cue"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &RootfsDo{})
}

type RootfsDoStepInterface struct {
	Input  Fs `json:"input,omitempty"`
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

	v := tt.Value().LookupPath(cue.ParsePath("input"))

	if err := v.Decode(&x.Input); err != nil {
		return err
	}

	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *dagger.Client) error {
		cc, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		step := &RootfsDoStepInterface{}
		step.Output = x.Input.Rootfs

		stepIter, err := cueflow.IterSteps(cueflow.CueValue(tt.Value()))
		if err != nil {
			return err
		}

		for idx, itemValue := range stepIter {
			path := cue.MakePath(slices.Concat(
				itemValue.Path().Selectors(),
				cue.ParsePath("input").Selectors(),
			)...)

			if err := tt.Scope().FillPath(path, step.Output); err != nil {
				return err
			}

			if err := tt.Scope().RunTasks(ctx, cueflow.WithPrefix(itemValue.Path())); err != nil {
				return fmt.Errorf("steps[%d]: %w", idx, err)
			}

			stepValue := tt.Scope().LookupPath(itemValue.Path())

			if err := stepValue.Decode(step); err != nil {
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
