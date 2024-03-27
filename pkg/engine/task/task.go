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

var _ cueflow.Group = &Group{}

type Group struct {
	Task

	parent cueflow.Task
}

func (v *Group) SetParent(t cueflow.Task) {
	v.parent = t
}

func (v *Group) Parent() cueflow.Task {
	return v.parent
}
