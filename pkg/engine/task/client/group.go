package client

import (
	"context"

	"cuelang.org/go/cue"
	"github.com/octohelm/cuekit/pkg/cuepath"

	"github.com/octohelm/cuekit/pkg/cueflow"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&GroupInterface{})
}

type GroupInterface struct {
	task.Group
}

func (x *GroupInterface) Do(ctx context.Context) error {
	tt := x.T()

	if err := cueflow.RunSubTasks(ctx, tt.Scope(), func(p cue.Path) (bool, cue.Path) {
		if cuepath.Prefix(p, tt.Path()) && !cuepath.Same(p, tt.Path()) {
			return true, cuepath.TrimPrefix(p, tt.Path())
		}
		return false, cue.MakePath()
	}); err != nil {
		return err
	}

	return nil
}
