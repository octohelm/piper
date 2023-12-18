package core

import "github.com/octohelm/piper/pkg/engine/plan/internal"

var DefaultFactory = internal.New()

type Task = internal.Task

type Result struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

type SetupTask struct {
	Task
}

func (v *SetupTask) Setup() bool {
	return true
}
