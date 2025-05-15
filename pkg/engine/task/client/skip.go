package client

import (
	"github.com/octohelm/cuekit/pkg/cueflow"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Skip{})
}

// Skip will skip task when matched
type Skip struct {
	cueflow.FlowControlImpl

	When bool `json:"when"`
}
