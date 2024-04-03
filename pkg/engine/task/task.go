package task

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/x/ptr"
)

type Task struct {
	// task hook to make task could run after some others
	Dep any `json:"$dep,omitempty"`
	// task result
	Ok *bool `json:"-" output:"$ok,omitempty"`

	cueflow.TaskImpl
}

var _ cueflow.Successor = &Task{}

func (t *Task) Success() bool {
	return t.Ok != nil && *t.Ok
}

var _ cueflow.TaskFeedback = &Task{}

func (t *Task) Done(err error) {
	if t.Ok == nil {
		t.Ok = ptr.Ptr(err == nil)
	}
}

type SetupTask struct {
	Task
}

var _ cueflow.TaskSetup = &SetupTask{}

func (SetupTask) Setup() bool {
	return true
}

type Group struct {
	Task

	t cueflow.Task
}

func (v *Group) T() cueflow.Task {
	return v.t
}

var _ cueflow.TaskUnmarshaler = &Group{}

func (v *Group) UnmarshalTask(t cueflow.Task) error {
	v.t = t
	return nil
}

var _ cueflow.CacheDisabler = &Group{}

func (Group) CacheDisabled() bool {
	return true
}

type Checkpoint struct {
	// no need the ok
	Task `json:"-"`
}

var _ cueflow.CacheDisabler = &Checkpoint{}

func (Checkpoint) CacheDisabled() bool {
	return true
}

var _ cueflow.Checkpoint = &Checkpoint{}

func (Checkpoint) AsCheckpoint() bool {
	return true
}

func (Checkpoint) Do(ctx context.Context) error {
	return nil
}
