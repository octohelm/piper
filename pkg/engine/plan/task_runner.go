package plan

import (
	"context"
	"cuelang.org/go/cue"
	contextx "github.com/octohelm/x/context"
)

var TaskRunnerFactoryContext = contextx.New[TaskRunnerFactory]()

type TaskRunner interface {
	Path() cue.Path
	Underlying() any
	Run(ctx context.Context) error
}

type StepRunner interface {
	Do(ctx context.Context) error
}

type TaskRunnerFactory interface {
	ResolveTaskRunner(task Task) (TaskRunner, error)
}
