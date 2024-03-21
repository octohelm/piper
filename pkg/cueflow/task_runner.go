package cueflow

import (
	"context"
	"log/slog"
	"reflect"

	"cuelang.org/go/cue"
	"github.com/go-courier/logr"
	contextx "github.com/octohelm/x/context"
	"github.com/pkg/errors"
)

var TaskRunnerFactoryContext = contextx.New[TaskRunnerResolver]()

type TaskRunnerResolver interface {
	ResolveTaskRunner(task Task) (TaskRunner, error)
}

type TaskRunner interface {
	Path() cue.Path
	Underlying() any
	Run(ctx context.Context) error
}

type StepRunner interface {
	Do(ctx context.Context) error
}

type WithScopeName interface {
	ScopeName(ctx context.Context) (string, error)
}

type taskRunner struct {
	task            Task
	inputTaskRunner reflect.Value
	outputFields    map[string]int
}

func (t *taskRunner) Underlying() any {
	return t.inputTaskRunner.Interface()
}

func (t *taskRunner) Path() cue.Path {
	return t.task.Path()
}

func (t *taskRunner) Task() Task {
	return t.task
}

func (t *taskRunner) resultValues() map[string]any {
	values := map[string]any{}

	rv := t.inputTaskRunner

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for name, i := range t.outputFields {
		if name == "" {
			f := rv.Field(i)
			if f.Kind() == reflect.Map {
				for _, k := range f.MapKeys() {
					key := k.String()
					if key == "$$task" {
						continue
					}
					values[key] = f.MapIndex(k).Interface()
				}
			}
			continue
		}
		values[name] = rv.Field(i).Interface()
	}

	return values
}

func (t *taskRunner) Run(ctx context.Context) (err error) {
	ctx = TaskPathContext.Inject(ctx, t.task.Path().String())

	taskValue := t.inputTaskRunner.Interface()

	stepRunner := taskValue.(StepRunner)
	l := logr.FromContext(ctx)

	if err := t.task.Decode(stepRunner); err != nil {
		return errors.Wrapf(err, "decode failed")
	}

	if n, ok := stepRunner.(WithScopeName); ok {
		scopeName, err := n.ScopeName(ctx)
		if err != nil {
			return err
		}
		l = l.WithValues(LogAttrScope, scopeName)
	}

	logAttrs := []any{
		slog.String(LogAttrName, t.task.Name()),
	}

	for _, d := range t.task.Deps() {
		logAttrs = append(logAttrs,
			slog.String(LogAttrDep, d.String()),
		)
	}

	ctx, l = l.Start(ctx, t.task.Path().String(), logAttrs...)
	defer l.End()

	l.WithValues("params", CueLogValue(stepRunner)).Debug("started.")

	if err := stepRunner.Do(ctx); err != nil {

		return errors.Wrapf(err, "%T do failed", stepRunner)
	}

	values := t.resultValues()

	defer func() {
		if err != nil {
			l.Error(err)
		} else {
			logTaskResult(l, stepRunner)
		}
	}()

	if err := t.task.Fill(values); err != nil {
		return errors.Wrap(err, "fill result failed")
	}

	return nil
}
