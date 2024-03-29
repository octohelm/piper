package cueflow

import (
	"context"
	"cuelang.org/go/cue"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/tools/flow"
	"github.com/pkg/errors"
)

type TaskOptionFunc func(c *flowTaskConfig)

type flowTaskConfig struct {
	prefix     *cue.Path
	shouldRun  func(value cue.Value) bool
	taskFunc   flow.RunnerFunc
	updateFunc func(c *flow.Controller, t *flow.Task) error
}

func (c *flowTaskConfig) Build(optFns ...TaskOptionFunc) {
	for _, optFn := range optFns {
		optFn(c)
	}

	if c.shouldRun == nil {
		c.shouldRun = func(value cue.Value) bool {
			return value.LookupPath(TaskPath).Exists()
		}
	}

	if c.taskFunc == nil {
		c.taskFunc = func(t *flow.Task) error {
			return nil
		}
	}
}

func isPrefixStrict(selectors []cue.Selector, prefixSelectors []cue.Selector) bool {
	if len(selectors) < len(prefixSelectors) {
		return false
	}

	for i, x := range prefixSelectors {
		if x.String() != selectors[i].String() {
			return false
		}
	}

	if len(selectors) > 0 {
		last := selectors[len(selectors)-1]
		if last.LabelType() != cue.IndexLabel {
			// struct path equal should not prefix
			if len(selectors) == len(prefixSelectors) {
				return false
			}
		}
	}

	return true
}

func trimPrefix(selectors []cue.Selector, prefixParts []cue.Selector) []cue.Selector {
	if isPrefixStrict(selectors, prefixParts) {
		return selectors[len(prefixParts):]
	}
	return selectors
}

func isInSlice(selectors []cue.Selector) bool {
	for _, x := range selectors {
		if x.Type() == cue.IndexLabel {
			return true
		}
	}
	return false
}

func (c *flowTaskConfig) New(v Scope) *flow.Controller {
	return flow.New(&flow.Config{
		FindHiddenTasks: true,
		UpdateFunc:      c.updateFunc,
	}, CueValue(v.Value()), func(v cue.Value) (flow.Runner, error) {
		selectors := v.Path().Selectors()

		if prefix := c.prefix; prefix != nil {
			if !isPrefixStrict(selectors, prefix.Selectors()) {
				return nil, nil
			}
		}

		if !(c.shouldRun(v)) {
			return nil, nil
		}

		if prefix := c.prefix; prefix != nil {
			if isInSlice(trimPrefix(selectors, prefix.Selectors())) {
				return nil, nil
			}
		} else {
			if isInSlice(selectors) {
				return nil, nil
			}
		}

		return c.taskFunc, nil
	})
}

func WithShouldRunFunc(shouldRun func(value cue.Value) bool) TaskOptionFunc {
	return func(c *flowTaskConfig) {
		c.shouldRun = shouldRun
	}
}

func WithPrefix(path cue.Path) TaskOptionFunc {
	return func(c *flowTaskConfig) {
		c.prefix = &path
	}
}

func resoleTasks(ctx context.Context, scope Scope, opts ...TaskOptionFunc) []*flow.Task {
	c := &flowTaskConfig{}
	c.Build(opts...)
	return c.New(scope).Tasks()
}

func runTasks(ctx context.Context, scope Scope, opts ...TaskOptionFunc) error {
	c := &flowTaskConfig{}
	c.Build(opts...)

	taskRunnerResolver := TaskRunnerFactoryContext.From(ctx)

	c.taskFunc = func(t *flow.Task) error {
		if scope.Processed(t.Path()) {
			return nil
		}

		tk := WrapTask(t, scope)

		tr, err := taskRunnerResolver.ResolveTaskRunner(tk)
		if err != nil {
			return errors.Wrap(err, "resolve task failed")
		}

		if err := tr.Run(ctx); err != nil {
			return cueerrors.Wrapf(err, tk.Value().Pos(), "%s run failed", tk.Name())
		}

		return nil
	}

	if err := c.New(scope).Run(ctx); err != nil {
		return err
	}

	return nil
}
