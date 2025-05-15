package internal

import (
	"context"
	"errors"
	"github.com/octohelm/cuekit/pkg/task"
	"os"
	"sync"

	"cuelang.org/go/cue"
)

var TaskPath = cue.ParsePath("$$task.name")

func New(v cue.Value, optionFuncs ...OptionFunc) *Controller {
	c := &Controller{}
	c.build(optionFuncs...)
	if err := c.init(v); err != nil {
		panic(err)
	}
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

func (x *Controller) nodeFromTask(t *task.Task) *node {
	if x.ranks == nil {
		x.ranks = map[string]int{}
	}

	x.ranks[t.Path().String()] += 1

	n := x.loadOrInit(t.Path())

	for d := range t.Deps() {
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

func (c *Controller) init(v cue.Value) error {
	ctrl := &task.Controller{
		IsTask: func(v cue.Value) bool {
			selectors := v.Path().Selectors()

			if prefix := c.Prefix; prefix != nil {
				if !isPrefixStrict(selectors, prefix.Selectors()) {
					return false
				}
			}

			return c.shouldRun(v)
		},
	}

	if err := ctrl.Init(v); err != nil {
		return err
	}

	for tt := range ctrl.Tasks() {
		_ = c.nodeFromTask(tt)
	}

	return nil
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
