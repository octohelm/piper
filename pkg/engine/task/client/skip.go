package client

import (
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Skip{})
}

// Skip will skip task when matched
type Skip struct {
	cueflow.FlowControlImpl

	When bool `json:"when"`
}
