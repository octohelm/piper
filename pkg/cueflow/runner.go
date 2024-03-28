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

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	cueerrors "cuelang.org/go/cue/errors"
	"cuelang.org/go/tools/flow"
	"github.com/gobwas/glob"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/vito/progrock"
	"github.com/vito/progrock/console"

	"github.com/octohelm/piper/pkg/dagger"
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
	build    func() (Value, error)
	root     atomic.Pointer[scope]
	taskDone sync.Map

	match  func(p string) bool
	output string

	setups     map[string][]string
	targets    map[string][]string
	graphPaths map[string][]string

	mu sync.RWMutex
}

func (r *Runner) printAllowedTasksTo(w io.Writer, tasks []*flow.Task) {
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
	p := formatPath(t.Path())
	if _, ok := collection[p]; ok {
		return
	}

	// avoid cycle
	collection[p] = make([]string, 0)

	deps := make([]string, 0)
	for _, d := range t.Dependencies() {
		deps = append(deps, formatPath(d.Path()))
		r.resolveDependencies(d, collection)
	}

	collection[p] = deps
}

func (r *Runner) Value() Value {
	return r.root.Load().Value
}

func (r *Runner) LookupPath(p cue.Path) Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.root.Load().Value.LookupPath(p)
}

func (r *Runner) FillPath(p cue.Path, v any) error {
	_, ok := r.taskDone.LoadOrStore(p.String(), true)
	if !ok {
		r.mu.Lock()
		defer r.mu.Unlock()

		r.root.Store(&scope{Value: r.root.Load().Value.FillPath(p, v)})
	}
	return nil
}

func (r *Runner) Processed(p cue.Path) bool {
	_, ok := r.taskDone.Load(p.String())
	return ok
}

func (r *Runner) RunTasks(ctx context.Context, optFns ...TaskOptionFunc) error {
	return runTasks(ctx, r, optFns...)
}

func (r *Runner) Run(ctx context.Context, action []string) error {
	actions := append([]string{"actions"}, action...)
	for i := range actions {
		actions[i] = strconv.Quote(actions[i])
	}

	p := cue.ParsePath(strings.Join(actions, "."))

	targetPath := formatPath(p)

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

func (r *Runner) init() error {
	rootValue, err := r.build()
	if err != nil {
		return err
	}
	r.root.Store(&scope{Value: rootValue})
	return nil
}

func (r *Runner) run(ctx context.Context) error {
	// init for scanning task
	if err := r.init(); err != nil {
		return err
	}

	if err := r.prepareTasks(ctx, resoleTasks(ctx, r)); err != nil {
		return errors.Wrap(err, "prepare task failed")
	}

	if len(r.targets) == 0 {
		return errors.New("no tasks founded to run")
	}

	if err := r.RunTasks(ctx, WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.setups[formatPath(value.Path())]
		return ok
	})); err != nil {
		return errors.Wrap(err, "run setup task failed")
	}

	if err := r.RunTasks(ctx, WithShouldRunFunc(func(value cue.Value) bool {
		_, ok := r.targets[formatPath(value.Path())]
		return ok
	})); err != nil {
		return errors.Wrap(err, "run task failed")
	}

	return nil
}

func formatPath(p cue.Path) string {
	b := &strings.Builder{}

	for i, s := range p.Selectors() {
		if i > 0 {
			b.WriteRune('/')
		}

		if s.Type() == cue.StringLabel {
			if strings.Contains(s.String(), "/") {
				b.WriteString(s.String())
				continue
			}
			b.WriteString(s.Unquoted())
			continue
		}
		b.WriteString(s.String())
	}

	return b.String()
}

func (r *Runner) walkTasks(ctx context.Context, tasks []*flow.Task, prefix []string) error {
	taskRunnerFactory := TaskRunnerFactoryContext.From(ctx)

	for i := range tasks {
		tk := WrapTask(tasks[i], r)

		taskPath := formatPath(tk.Path())

		prefixFullPath := strings.Join(prefix, "/")

		var currentPath = prefix

		if strings.HasPrefix(taskPath, prefixFullPath) && taskPath != prefixFullPath {
			currentPath = append(prefix, strings.TrimPrefix(taskPath, prefixFullPath+"/"))
		}

		r.graphPaths[taskPath] = currentPath

		t, err := taskRunnerFactory.ResolveTaskRunner(tk)
		if err != nil {
			return cueerrors.Wrapf(err, tk.Value().Pos(), "resolve task failed")
		}

		switch t.Underlying().(type) {
		case TaskSetup:
			r.resolveDependencies(tasks[i], r.setups)
		case TaskUnmarshaler:
			if r.match(formatPath(tk.Path())) {
				stepIter, err := IterSteps(CueValue(tk.Value()))
				if err == nil {
					for i, step := range stepIter {
						stepPath := append(currentPath, fmt.Sprintf("steps/%d", i))

						if i > 0 {
							// dep pre step
							r.targets[formatPath(step.Path())] = []string{
								strings.Join(append(currentPath, fmt.Sprintf("steps/%d", i-1)), "/"),
							}

							r.graphPaths[formatPath(step.Path())] = stepPath
						}

						if err := r.walkTasks(ctx, resoleTasks(ctx, r, WithPrefix(step.Path())), stepPath); err != nil {
							return err
						}
					}
				}
			}
		}

		if r.match(formatPath(tk.Path())) {
			r.resolveDependencies(tasks[i], r.targets)
		}
	}

	return nil
}

func (r *Runner) prepareTasks(ctx context.Context, tasks []*flow.Task) error {
	r.graphPaths = map[string][]string{}
	r.setups = map[string][]string{}
	r.targets = map[string][]string{}

	if err := r.walkTasks(ctx, tasks, nil); err != nil {
		return err
	}

	if len(r.targets) > 0 {
		if os.Getenv("GRAPH") != "" {
			fmt.Println(printGraph(r.targets, r.graphPaths))
		}
		return nil
	}

	buf := bytes.NewBuffer(nil)
	r.printAllowedTasksTo(buf, tasks)
	return errors.New(buf.String())
}

func printGraph(targets map[string][]string, graphPaths map[string][]string) (string, error) {
	buffer := bytes.NewBuffer(nil)

	w, err := zlib.NewWriterLevel(buffer, 9)
	if err != nil {
		return "", errors.Wrap(err, "fail to create the w")
	}

	wrap := func(name string) string {
		if parts, ok := graphPaths[name]; ok {
			p := make([]string, len(parts))
			for i, x := range parts {
				p[i] = fmt.Sprintf("%q", x)
			}
			return strings.Join(p, ".")
		}
		return fmt.Sprintf("%q", name)
	}

	_, _ = fmt.Fprintf(w, `direction: right
`)
	for name, deps := range targets {
		for _, d := range deps {
			_, _ = fmt.Fprintf(w, `
%s -> %s
`, wrap(d), wrap(name))
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
