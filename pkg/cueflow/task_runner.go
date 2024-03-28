package cueflow

import (
	"context"
	"encoding/json"
	"github.com/opencontainers/go-digest"
	"log/slog"
	"reflect"
	"sync"

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
	outputFields    outputFields
	inputTaskRunner reflect.Value
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

func (t *taskRunner) Run(ctx context.Context) (err error) {
	ctx = TaskPathContext.Inject(ctx, t.task.Path().String())

	stepRunner := t.inputTaskRunner.Interface().(StepRunner)

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

	output, err := t.cachedOrDoTask(ctx, stepRunner)
	if err != nil {
		return err
	}

	// fill output for tasks at flow.*
	if taskFeedback, ok := stepRunner.(TaskFeedback); ok {
		taskFeedback.FillResult(output)
	}

	// fill output to cue
	if err := t.task.Fill(output); err != nil {
		return errors.Wrap(err, "fill result failed")
	}

	return nil
}

var cache = &sync.Map{}

type cacheKey struct {
	name string
	hash string
}

func (t *taskRunner) cachedOrDoTask(ctx context.Context, stepRunner StepRunner) (output map[string]any, err error) {
	l := logr.FromContext(ctx)
	hint := false

	defer func() {
		done := "done."
		if hint {
			done = "cached."
		}

		if err != nil {
			l.Error(err)
		} else {
			keyAndValues := make([]any, 0, len(output)*2)
			for k, v := range output {
				keyAndValues = append(keyAndValues, k, CueLogValue(v))
			}
			l.WithValues(keyAndValues...).Debug(done)
		}
	}()

	params, err := json.Marshal(stepRunner)
	if err != nil {
		return nil, err
	}

	do := sync.OnceValue(func() any {
		l.WithValues("params", params).Debug("started.")

		if err := stepRunner.Do(ctx); err != nil {
			return errors.Wrapf(err, "%T do failed", stepRunner)
		}

		// done task before resolve output values
		if taskFeedback, ok := stepRunner.(TaskFeedback); ok {
			taskFeedback.Done(nil)
		}

		return t.outputFields.OutputValues(t.inputTaskRunner)
	})

	if cacheDisabler, ok := stepRunner.(CacheDisabler); ok && cacheDisabler.CacheDisabled() {
		//
	} else {
		key := cacheKey{
			name: t.task.Name(),
			hash: digest.FromBytes(params).String(),
		}
		actual, ok := cache.LoadOrStore(key, do)
		do, hint = actual.(func() any), ok
	}

	switch x := do().(type) {
	case error:
		return nil, x
	case map[string]any:
		return x, nil
	default:
		return map[string]any{}, nil
	}
}

type outputFields map[string][]int

func (fields outputFields) OutputValues(rv reflect.Value) map[string]any {
	if valuer, ok := rv.Interface().(OutputValuer); ok {
		return valuer.OutputValues()
	}

	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	values := map[string]any{}

	for name, loc := range fields {
		f := getField(rv, loc)

		if name == "" {
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
		values[name] = f.Interface()
	}

	return values
}

func getField(rv reflect.Value, loc []int) reflect.Value {
	switch len(loc) {
	case 0:
		return rv
	case 1:
		return rv.Field(loc[0])
	default:
		return getField(rv.Field(loc[0]), loc[1:])
	}
}
