package cueflow

import "cuelang.org/go/cue"

var (
	DepPath     = cue.ParsePath("$dep")
	OkPath      = cue.ParsePath("$ok")
	ControlPath = cue.ParsePath("$$control.name")
)

type FlowTask interface {
	flowTask()
}

type FlowControl interface {
	FlowTask

	flowControl()
}

// TaskSetup which task will run before all others tasks
type TaskSetup interface {
	Setup() bool
}

type TaskImpl struct{}

func (TaskImpl) flowTask() {
}

var _ FlowControl = FlowControlImpl{}

type FlowControlImpl struct{}

func (FlowControlImpl) flowTask() {
}

func (FlowControlImpl) flowControl() {
}

type OutputValuer interface {
	OutputValues() map[string]any
}

type CanSkip interface {
	Skip() bool
}

type CacheDisabler interface {
	CacheDisabled() bool
}

type Checkpoint interface {
	AsCheckpoint() bool
}

type Successor interface {
	Success() bool
}

type ResultValuer interface {
	ResultValue() map[string]any
}

type TaskFeedback interface {
	Done(err error)
}
