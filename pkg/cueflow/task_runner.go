package cueflow

import (
	"context"
	encodingcue "github.com/octohelm/piper/pkg/encoding/cue"
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
	stepRunner := t.inputTaskRunner.Interface().(StepRunner)

	cv := CueValue(t.task.Value())
	dep := cv.LookupPath(DepPath)
	if ctrl := dep.LookupPath(ControlPath); ctrl.Exists() {
		ctrlType, _ := ctrl.String()
		switch ctrlType {
		case "skip":
			needSkip, _ := dep.LookupPath(cue.ParsePath("when")).Bool()
			if needSkip {
				if _, ok := stepRunner.(TaskFeedback); ok {
					return t.fill(map[string]any{
						"$ok": false,
					})
				}
				return nil
			}
		}
	}

	ctx = TaskPathContext.Inject(ctx, t.task.Path().String())

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

	return t.fill(output)
}

func (t *taskRunner) fill(output map[string]any) error {
	// fill output to root value
	if err := t.task.Scope().FillPath(t.task.Path(), output); err != nil {
		return errors.Wrap(err, "fill result values failed")
	}
	// fill output to trigger cue flow continue
	if err := t.task.Fill(nil); err != nil {
		return errors.Wrap(err, "fill task failed")
	}

	return nil
}

var cache = &sync.Map{}

type cacheKey struct {
	name string
	hash string
}

func (t *taskRunner) cachedOrDoTask(ctx context.Context, stepRunner StepRunner) (output map[string]any, err error) {
	isCheckpoint := false

	if checkpoint, ok := stepRunner.(Checkpoint); ok {
		isCheckpoint = checkpoint.AsCheckpoint()
	}

	l := logr.FromContext(ctx)
	hint := false

	defer func() {
		if err != nil {
			l.Error(err)
			return
		}
		done := "done"
		if hint {
			done = "cached"
		}
		if isCheckpoint {
			done = "resolved"
		}
		l.WithValues("result", CueLogValue(output)).Debug(done)
	}()

	params, err := encodingcue.Marshal(stepRunner)
	if err != nil {
		return nil, err
	}

	do := sync.OnceValue(func() any {
		if !isCheckpoint {
			l.WithValues("params", params).Debug("starting")

			if err := stepRunner.Do(ctx); err != nil {
				return errors.Wrapf(err, "%T do failed", stepRunner)
			}
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

		// nil value never as output value
		if f.Kind() == reflect.Ptr {
			if !f.IsNil() {
				values[name] = f.Interface()
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
