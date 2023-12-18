package cueflow

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"

	"github.com/mattn/go-isatty"
	"github.com/vito/progrock"
	"github.com/vito/progrock/console"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/tools/flow"
	"github.com/go-courier/logr"
	"github.com/pkg/errors"
)

func NewRunner(value Value) *Runner {
	r := &Runner{}
	r.root.Store(&scope{Value: value})
	return r
}

type scope struct {
	Value Value
}

type Runner struct {
	root     atomic.Pointer[scope]
	taskDone sync.Map

	target cue.Path
	output string

	setups  map[string][]string
	targets map[string][]string
}

func (r *Runner) printAllowedTasksTo(w io.Writer, tasks []*flow.Task) {
	scope := r.target

	_, _ = fmt.Fprintf(w, `
Undefined action:

`)
	printSelectors(w, scope.Selectors()[1:]...)

	_, _ = fmt.Fprintf(w, `
Allowed action:

`)

	taskSelectors := map[string][]cue.Selector{}

	for _, t := range tasks {
		selectors := t.Path().Selectors()

		if selectors[0].String() == "actions" {
			publicSelectors := make([]cue.Selector, 0, len(selectors)-1)

			func() {
				for _, selector := range selectors[1:] {
					if selector.String()[0] == '_' {
						return
					}
					publicSelectors = append(publicSelectors, selector)
				}
			}()

			taskSelectors[cue.MakePath(publicSelectors...).String()] = publicSelectors
		}
	}

	keys := make([]string, 0, len(taskSelectors))
	for k := range taskSelectors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)
	defer func() {
		_ = tw.Flush()
	}()

	for _, k := range keys {
		printSelectors(tw, taskSelectors[k]...)

		taskValue := r.Value().(CueValueWrapper).CueValue().LookupPath(cue.ParsePath("actions." + k))

		if n := taskValue.Source(); n != nil {
			for _, c := range ast.Comments(n) {
				_, _ = fmt.Fprintf(tw, "\t\t%s", strings.TrimSpace(c.Text()))
			}
		}

		_, _ = fmt.Fprintln(tw)
	}
}

func printSelectors(w io.Writer, selectors ...cue.Selector) {
	for i, s := range selectors {
		if i > 0 {
			_, _ = fmt.Fprintf(w, ` `)
		}
		_, _ = fmt.Fprintf(w, `%s`, s.String())
	}
}

func (r *Runner) resolveDependencies(t *flow.Task, collection map[string][]string) {
	p := t.Path().String()
	if _, ok := collection[p]; ok {
		return
	}

	// avoid cycle
	collection[p] = make([]string, 0)

	deps := make([]string, 0)
	for _, d := range t.Dependencies() {
		deps = append(deps, d.Path().String())
		r.resolveDependencies(d, collection)
	}

	collection[p] = deps
}

func (r *Runner) Value() Value {
	return r.root.Load().Value
}

func (r *Runner) Fill(p cue.Path, v Value) error {
	r.taskDone.Store(p.String(), true)
	r.root.Store(&scope{Value: r.Value().FillPath(p, v)})
	return nil
}

func (r *Runner) Processed(p cue.Path) bool {
	_, ok := r.taskDone.Load(p.String())
	return ok
}

func (r *Runner) Run(ctx context.Context, action []string) error {
	actions := append([]string{"actions"}, action...)
	for i := range actions {
		actions[i] = strconv.Quote(actions[i])
	}

	r.target = cue.ParsePath(strings.Join(actions, "."))

	return runWith(ctx, r.target.String(), func(ctx context.Context) error {
		return r.run(ctx)
	})
}

func (r *Runner) run(ctx context.Context) error {
	f := NewFlow(r, noOpRunner)

	if err := r.prepareTasks(ctx, f.Tasks()); err != nil {
		return err
	}

	defer func() {
		if o := r.Value().LookupPath(r.target).LookupPath(cue.ParsePath("result")); o.Exists() {
			_, final := logr.FromContext(ctx).Start(ctx, o.Path().String())
			defer final.End()
			if ok := o.LookupPath(cue.ParsePath("ok")); ok.Exists() {
				ok, _ := CueValue(ok).Bool()
				if ok {
					final.WithValues("result", CueLogValue(o)).Info("success.")
				} else {
					final.WithValues("result", CueLogValue(o)).Error(errors.New("failed."))
				}
			} else {
				final.WithValues("result", CueLogValue(o)).Info("done.")
			}
		}
	}()

	if err := RunTasks(ctx, r, WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.setups[value.Path().String()]
		return ok
	})); err != nil {
		return err
	}

	if err := RunTasks(ctx, r, WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.targets[value.Path().String()]
		return ok
	})); err != nil {
		return err
	}

	return nil
}

func (r *Runner) prepareTasks(ctx context.Context, tasks []*flow.Task) error {
	taskRunnerFactory := TaskRunnerFactoryContext.From(ctx)

	r.setups = map[string][]string{}
	r.targets = map[string][]string{}

	for i := range tasks {
		tk := WrapTask(tasks[i], r)

		t, err := taskRunnerFactory.ResolveTaskRunner(tk)
		if err != nil {
			return cueerrors.Wrapf(err, tk.Value().Pos(), "resolve task failed")
		}

		if _, ok := t.Underlying().(interface{ Setup() bool }); ok {
			r.resolveDependencies(tasks[i], r.setups)
		}

		if strings.HasPrefix(tk.Path().String(), r.target.String()) {
			r.resolveDependencies(tasks[i], r.targets)
		}
	}

	if r.target.String() != "actions" && len(r.targets) > 0 {
		if os.Getenv("GRAPH") != "" {
			fmt.Println(printGraph(r.targets))
		}
		return nil
	}

	buf := bytes.NewBuffer(nil)
	r.printAllowedTasksTo(buf, tasks)
	return errors.New(buf.String())
}

func noOpRunner(cueValue cue.Value) (flow.Runner, error) {
	v := cueValue.LookupPath(TaskPath)

	if !v.Exists() {
		return nil, nil
	}

	// task in slice not be valid task
	for _, s := range v.Path().Selectors() {
		if s.Type() == cue.IndexLabel {
			return nil, nil
		}
	}

	return flow.RunnerFunc(func(t *flow.Task) error {
		return nil
	}), nil
}

func printGraph(targets map[string][]string) (string, error) {
	buffer := bytes.NewBuffer(nil)

	w, err := zlib.NewWriterLevel(buffer, 9)
	if err != nil {
		return "", errors.Wrap(err, "fail to create the w")
	}

	_, _ = fmt.Fprintf(w, "direction: right\n")
	for name, deps := range targets {
		for _, d := range deps {
			_, _ = fmt.Fprintf(w, "%q -> %q\n", d, name)
		}
	}
	_ = w.Close()
	if err != nil {
		return "", errors.Wrap(err, "fail to create the payload")
	}
	return fmt.Sprintf("https://kroki.io/d2/svg/%s?theme=101", base64.URLEncoding.EncodeToString(buffer.Bytes())), nil
}

var isTTY = sync.OnceValue(func() bool {
	if os.Getenv("TTY") == "0" {
		return false
	}
	// ugly to make as non-tty for Run of intellij
	if os.Getenv("_INTELLIJ_FORCE_PREPEND_PATH") != "" {
		return false
	}

	for _, f := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
		if isatty.IsTerminal(f.Fd()) {
			return true
		}
	}
	return false
})

func runWith(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	if isTTY() {
		tape := progrock.NewTape()
		tape.ShowInternal(true)
		tape.ShowAllOutput(true)
		tape.VerboseEdges(true)

		rec := progrock.NewRecorder(tape)

		ctx = progrock.ToContext(ctx, rec)

		defer func() {
			_ = rec.Close()

			if e := recover(); e != nil {
				fmt.Printf("%#v", e)
			}
		}()

		return progrock.DefaultUI().Run(ctx, tape, func(ctx context.Context, client progrock.UIClient) error {
			client.SetStatusInfo(progrock.StatusInfo{
				Name:  "Action",
				Value: name,
				Order: 1,
			})
			return fn(ctx)
		})
	}

	rootRec := progrock.NewRecorder(
		console.NewWriter(os.Stdout,
			console.ShowInternal(true),
		),
	)

	return fn(progrock.ToContext(ctx, rootRec))
}
