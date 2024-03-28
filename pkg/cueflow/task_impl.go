package cueflow

type FlowTask interface {
	flowTask()
}

// TaskSetup which task will run before all others tasks
type TaskSetup interface {
	Setup() bool
}

type TaskImpl struct {
}

func (TaskImpl) flowTask() {
}

type OutputValuer interface {
	OutputValues() map[string]any
}

type CacheDisabler interface {
	CacheDisabled() bool
}

type Successor interface {
	Success() bool
}

type ResultValuer interface {
	ResultValue() map[string]any
}

type TaskFeedback interface {
	Done(err error)
	FillResult(values map[string]any)
}
