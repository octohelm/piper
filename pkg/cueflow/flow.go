package cueflow

import (
	"cuelang.org/go/tools/flow"
)

func NewFlow(v Scope, taskFunc flow.TaskFunc) *flow.Controller {
	return flow.New(&flow.Config{
		FindHiddenTasks: true,
		UpdateFunc: func(c *flow.Controller, t *flow.Task) error {
			if t != nil {
				// when task value changes
				// need to put back value to root for using by child tasks
				return v.Fill(t.Path(), WrapValue(t.Value()))
			}
			return nil
		},
	}, CueValue(v.Value()), taskFunc)
}
