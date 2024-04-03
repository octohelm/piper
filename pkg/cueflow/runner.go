package cueflow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	cueerrors "cuelang.org/go/cue/errors"
	"github.com/gobwas/glob"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/vito/progrock"
	"github.com/vito/progrock/console"

	"github.com/octohelm/piper/pkg/cueflow/internal"
	"github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/generic/record"
)

func NewRunner(build func() (Value, error)) *Runner {
	return &Runner{
		build: build,
	}
}

type scope struct {
	Value Value
}

type Runner struct {
	build      func() (Value, error)
	root       atomic.Pointer[scope]
	taskResult record.Map[string, any]

	match  func(p string) bool
	output string

	mu sync.RWMutex

	setups        map[string][]string
	targets       map[string][]string
	activeTargets map[string][]string
}

func (r *Runner) Value() Value {
	return r.root.Load().Value
}

func (r *Runner) LookupResult(p cue.Path) (any, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.taskResult.Load(internal.FormatAsJSONPath(p))
}

func (r *Runner) LookupPath(p cue.Path) Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.root.Load().Value.LookupPath(p)
}

func (r *Runner) FillPath(p cue.Path, v any) error {
	if _, ok := v.(cue.Value); ok {
		return errors.Errorf("invalid value for filling %s", p)
	}

	_, ok := r.taskResult.LoadOrStore(internal.FormatAsJSONPath(p), v)
	if !ok {
		r.mu.Lock()
		defer r.mu.Unlock()

		r.root.Store(&scope{Value: r.root.Load().Value.FillPath(p, v)})
	}
	return nil
}

func (r *Runner) Processed(p cue.Path) bool {
	_, processed := r.taskResult.Load(internal.FormatAsJSONPath(p))
	return processed
}

func (r *Runner) RunTasks(ctx context.Context, optFns ...TaskOptionFunc) error {
	taskRunnerResolver := TaskRunnerFactoryContext.From(ctx)

	return internal.New(CueValue(r.Value()), append(optFns, internal.WithRunTask(func(ctx context.Context, n internal.Node) error {
		if r.Processed(n.Path()) {
			return nil
		}

		tk := NewTask(r, n)

		tr, err := taskRunnerResolver.ResolveTaskRunner(tk)
		if err != nil {
			return errors.Wrap(err, "resolve task failed")
		}

		if err := tr.Run(ctx); err != nil {
			return cueerrors.Wrapf(err, tk.Value().Pos(), "%s run failed", tk.Name())
		}

		return nil
	}))...).Run(ctx)
}

func (r *Runner) Run(ctx context.Context, action []string) error {
	actions := append([]string{"actions"}, action...)
	for i := range actions {
		actions[i] = strconv.Quote(actions[i])
	}

	p := cue.ParsePath(strings.Join(actions, "."))

	targetPath := internal.FormatAsJSONPath(p)

	g := glob.MustCompile(targetPath)

	r.match = func(taskPath string) bool {
		if strings.HasPrefix(taskPath, targetPath) {
			return true
		}
		return g.Match(taskPath)
	}

	return runWith(ctx, targetPath, func(ctx context.Context) error {
		return r.run(ctx)
	})
}

func (r *Runner) run(ctx context.Context) error {
	if err := r.init(); err != nil {
		return err
	}

	if err := r.scanTargets(ctx, internal.New(CueValue(r.Value()))); err != nil {
		return errors.Wrap(err, "prepare task failed")
	}

	if err := r.RunTasks(ctx, internal.WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.setups[internal.FormatAsJSONPath(value.Path())]
		return ok
	})); err != nil {
		return errors.Wrap(err, "run setup task failed")
	}

	if err := r.RunTasks(ctx, internal.WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.activeTargets[internal.FormatAsJSONPath(value.Path())]
		return ok
	})); err != nil {
		return errors.Wrap(err, "run task failed")
	}

	return nil
}

func (r *Runner) init() error {
	rootValue, err := r.build()
	if err != nil {
		return err
	}
	r.root.Store(&scope{Value: rootValue})
	return nil
}

func (r *Runner) resolveDependencies(t internal.Node, collection map[string][]string) {
	p := t.String()

	if _, ok := collection[p]; ok {
		return
	}

	// avoid cycle
	collection[p] = make([]string, 0)

	depNodes := t.Deps()
	deps := make([]string, 0, len(depNodes))
	for _, d := range depNodes {
		deps = append(deps, d.String())

		r.resolveDependencies(d, collection)
	}

	collection[p] = deps
}

func (r *Runner) printAllowedTasksTo(w io.Writer, tasks []internal.Node) {
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

func (r *Runner) walkTasks(ctx context.Context, tasks []internal.Node, prefix []string) error {
	taskRunnerFactory := TaskRunnerFactoryContext.From(ctx)

	for i, tk := range tasks {
		t, err := taskRunnerFactory.ResolveTaskRunner(NewTask(r, tk))
		if err != nil {
			return errors.New("resolve task failed")
		}

		switch t.Underlying().(type) {
		case TaskSetup:
			r.resolveDependencies(tasks[i], r.setups)
		}

		r.resolveDependencies(tasks[i], r.targets)
	}

	return nil
}

func (r *Runner) scanTargets(ctx context.Context, c *internal.Controller) error {
	r.setups = map[string][]string{}
	r.targets = map[string][]string{}

	if err := r.walkTasks(ctx, c.Tasks(), nil); err != nil {
		return err
	}

	r.scanActiveTarget()

	if len(r.activeTargets) > 0 {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	r.printAllowedTasksTo(buf, c.Tasks())
	return errors.New(buf.String())
}

func (r *Runner) scanActiveTarget() {
	r.activeTargets = map[string][]string{}

	var walkActiveTask func(name string)
	walkActiveTask = func(name string) {
		// avoid loop
		if _, ok := r.activeTargets[name]; ok {
			return
		}

		if deps, ok := r.targets[name]; ok {
			r.activeTargets[name] = deps
			for _, dep := range deps {
				walkActiveTask(dep)
			}
		}
	}

	for path, deps := range r.targets {
		if r.match(path) {
			walkActiveTask(path)
			for _, dep := range deps {
				walkActiveTask(dep)
			}
		}
	}
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

var debugEnabled = false

func init() {
	if os.Getenv("DEBUG") != "" {
		debugEnabled = true
	}
}

func runWith(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	if isTTY() {
		tape := progrock.NewTape()
		tape.ShowInternal(debugEnabled)
		tape.ShowAllOutput(true)
		tape.VerboseEdges(true)

		rec := progrock.NewRecorder(WrapProgrockWriter(tape))

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
			r, err := dagger.NewRunner(dagger.WithProgrockWriter(WrapProgrockWriter(tape)))
			if err != nil {
				return err
			}
			daggerRunner := WrapDaggerRunner(r)
			ctx = dagger.RunnerContext.Inject(ctx, daggerRunner)
			return fn(progrock.ToContext(ctx, rec))
		})
	}

	w := WrapProgrockWriter(console.NewWriter(os.Stdout, console.ShowInternal(debugEnabled)))
	r, err := dagger.NewRunner(dagger.WithProgrockWriter(WrapProgrockWriter(w)))
	if err != nil {
		return err
	}
	daggerRunner := WrapDaggerRunner(r)
	ctx = dagger.RunnerContext.Inject(ctx, daggerRunner)
	return fn(progrock.ToContext(ctx, progrock.NewRecorder(w)))
}
