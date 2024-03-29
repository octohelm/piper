package client

import (
	"context"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &GroupInterface{})
}

// Group
type GroupInterface struct {
	task.Group
}

func (x *GroupInterface) Do(ctx context.Context) error {
	tt := x.T()

	if err := tt.Scope().RunTasks(ctx, cueflow.WithPrefix(tt.Path())); err != nil {
		return err
	}

	return nil
}
