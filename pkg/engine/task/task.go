package task

import "github.com/octohelm/piper/pkg/cueflow"

type Task = cueflow.TaskImpl

type SetupTask struct {
	Task

	cueflow.IsSetup
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
