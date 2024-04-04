package internal

import (
	"context"
	"github.com/pkg/errors"
	"os"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
)

var TaskPath = cue.ParsePath("$$task.name")

func New(v cue.Value, optionFuncs ...OptionFunc) *Controller {
	c := &Controller{}
	c.build(optionFuncs...)
	c.init(v)
	return c
}

type Controller struct {
	runTask   func(ctx context.Context, n Node) error
	shouldRun func(value cue.Value) bool
	Prefix    *cue.Path

	nodes   map[string]*node
	ranks   map[string]int
	taskFns sync.Map
}

func (c *Controller) Tasks() []Node {
	nodes := make([]Node, 0, len(c.nodes))
	for _, n := range c.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

type OptionFunc func(c *Controller)

func WithShouldRunFunc(shouldRun func(value cue.Value) bool) OptionFunc {
	return func(c *Controller) {
		c.shouldRun = shouldRun
	}
}

func WithRunTask(runTask func(ctx context.Context, n Node) error) OptionFunc {
	return func(c *Controller) {
		c.runTask = runTask
	}
}

func (c *Controller) build(optFns ...OptionFunc) {
	for _, optFn := range optFns {
		optFn(c)
	}

	if c.shouldRun == nil {
		c.shouldRun = func(value cue.Value) bool {
			return value.LookupPath(TaskPath).Exists()
		}
	}
}

func (x *Controller) runOnce(ctx context.Context, n Node) error {
	once, _ := x.taskFns.LoadOrStore(n.Path().String(), sync.OnceValue(func() error {
		return x.runTask(ctx, n)
	}))
	return once.(func() error)()
}

var showGraph = false

func init() {
	showGraph = os.Getenv("GRAPH") != ""
}

func (x *Controller) Run(ctx context.Context) error {
	if showGraph {
		if prefix := x.Prefix; prefix != nil {
			printGraph(prefix.String(), x.Tasks())
		} else {
			printGraph("_", x.Tasks())
		}
	}

	eg, c := WithContext(ctx)
	for p, rank := range x.ranks {
		// trigger the final task
		if rank == 1 {
			n := x.nodes[p]
			eg.Go(func() error {
				return n.Run(c)
			})
		}
	}

	if err := eg.Wait(); err != nil {
		// ignore cancel
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return nil
}

func (x *Controller) nodeFromTask(t *flow.Task) *node {
	if x.ranks == nil {
		x.ranks = map[string]int{}
	}

	x.ranks[t.Path().String()] += 1

	n := x.loadOrInit(t.Path())

	for _, d := range t.Dependencies() {
		n.addDep(x.nodeFromTask(d))
	}

	return n
}

func (x *Controller) loadOrInit(p cue.Path) *node {
	if x.nodes == nil {
		x.nodes = map[string]*node{}
	}
	if found, ok := x.nodes[p.String()]; ok {
		return found
	}
	created := &node{path: p, Controller: x}
	x.nodes[p.String()] = created
	return created

}

func (c *Controller) init(v cue.Value) {
	fc := &flow.Config{
		FindHiddenTasks: true,
	}

	ctrl := flow.New(fc, v, func(v cue.Value) (flow.Runner, error) {
		selectors := v.Path().Selectors()

		if prefix := c.Prefix; prefix != nil {
			if !isPrefixStrict(selectors, prefix.Selectors()) {
				return nil, nil
			}
		}

		if !(c.shouldRun(v)) {
			return nil, nil
		}

		if prefix := c.Prefix; prefix != nil {
			if isInCueSlice(trimPrefix(selectors, prefix.Selectors())) {
				return nil, nil
			}
		} else {
			if isInCueSlice(selectors) {
				return nil, nil
			}
		}

		return flow.RunnerFunc(func(t *flow.Task) error {
			// do nothing
			// just use for task resolver
			return nil
		}), nil
	})

	for _, tt := range ctrl.Tasks() {
		_ = c.nodeFromTask(tt)
	}
}

func trimPrefix(selectors []cue.Selector, prefixParts []cue.Selector) []cue.Selector {
	if isPrefixStrict(selectors, prefixParts) {
		return selectors[len(prefixParts):]
	}
	return selectors
}

func isInCueSlice(selectors []cue.Selector) bool {
	for _, x := range selectors {
		if x.Type() == cue.IndexLabel {
			return true
		}
	}
	return false
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
