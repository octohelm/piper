package cueflow

import (
	"context"

	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/tools/flow"
	"github.com/pkg/errors"
)

type RunTaskOptionFunc func(c *taskController)

func RunTasks(ctx context.Context, scope Scope, opts ...RunTaskOptionFunc) error {
	tr := &taskController{
		scope:              scope,
		taskRunnerResolver: TaskRunnerFactoryContext.From(ctx),
		shouldRun: func(value cue.Value) bool {
			return value.LookupPath(TaskPath).Exists()
		},
	}

	tr.Build(opts...)

	if err := tr.Run(ctx, scope); err != nil {
		return err
	}

	return nil
}

func WithShouldRunFunc(shouldRun func(value cue.Value) bool) RunTaskOptionFunc {
	return func(c *taskController) {
		c.shouldRun = shouldRun
	}
}

func WithPrefix(path cue.Path) RunTaskOptionFunc {
	return func(c *taskController) {
		c.prefix = path
	}
}

type taskController struct {
	scope              Scope
	taskRunnerResolver TaskRunnerResolver
	shouldRun          func(value cue.Value) bool
	prefix             cue.Path
}

func (fc *taskController) Build(optFns ...RunTaskOptionFunc) {
	for _, optFn := range optFns {
		optFn(fc)
	}
}

func (fc *taskController) Run(ctx context.Context, scope Scope) error {
	return NewFlow(scope, func(cueValue cue.Value) (flow.Runner, error) {
		if !(fc.shouldRun(cueValue)) {
			return nil, nil
		}

		return flow.RunnerFunc(func(t *flow.Task) (err error) {
			if fc.scope.Processed(t.Path()) {
				return nil
			}

			tk := WrapTask(t, scope)

			tr, err := fc.taskRunnerResolver.ResolveTaskRunner(tk)
			if err != nil {
				return errors.Wrap(err, "resolve task failed")
			}

			if err := tr.Run(ctx); err != nil {
				return cueerrors.Wrapf(err, tk.Value().Pos(), "%s run failed", tk.Name())
			}

			return nil
		}), nil
	}).Run(ctx)
}
