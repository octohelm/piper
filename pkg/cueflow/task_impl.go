package cueflow

func RegisterTask(r TaskImplRegister, task FlowTask) {
	r.Register(task)
}

type TaskImplRegister interface {
	Register(t any)
}

type FlowTask interface {
	flowTask()
}

type TaskImpl struct {
}

func (TaskImpl) flowTask() {
}

type IsSetup interface {
	Setup() bool
}
