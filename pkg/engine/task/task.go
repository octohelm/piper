package task

import (
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/x/ptr"
)

type Task struct {
	// task result
	Ok *bool `json:"-" output:"ok"`

	values map[string]any

	cueflow.TaskImpl
}

var _ cueflow.TaskFeedback = &Task{}

func (t *Task) Done(err error) {
	if t.Ok == nil {
		t.Ok = ptr.Ptr(err == nil)
	}
}

func (t *Task) Success() bool {
	return t.Ok != nil && *t.Ok
}

func (t *Task) SetResultValue(v map[string]any) {
	t.values = v
}

var _ cueflow.ResultValuer = Task{}

func (t Task) ResultValue() map[string]any {
	return t.values
}

type SetupTask struct {
	Task
}

var _ cueflow.IsSetup = &SetupTask{}

func (SetupTask) Setup() bool {
	return true
}

var _ cueflow.TaskUnmarshaler = &Group{}

type Group struct {
	Task

	t cueflow.Task
}

func (v *Group) UnmarshalTask(t cueflow.Task) error {
	v.t = t
	return nil
}

func (v *Group) T() cueflow.Task {
	return v.t
}
