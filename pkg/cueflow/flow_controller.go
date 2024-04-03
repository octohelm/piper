package cueflow

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/tools/flow"
	"github.com/pkg/errors"
	"sync"
)

type controller struct {
	nodes   map[string]*node
	ranks   map[string]int
	started sync.Map

	runTask func(ctx context.Context, n Node) error
}

func (x *controller) runOnce(ctx context.Context, n Node) error {
	once, _ := x.started.LoadOrStore(n.Path().String(), sync.OnceValue(func() error {
		return x.runTask(ctx, n)
	}))

	return once.(func() error)()
}

func (x *controller) Run(ctx context.Context) error {
	errOnce := NewErrOnce()

	for path, rank := range x.ranks {
		// trigger the final task
		if rank == 1 {
			n := x.nodes[path]

			errOnce.Go(func() error {
				return n.Run(ctx)
			})
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-errOnce.Done():
		if err := errOnce.Err(); err != nil {
			return err
		}
	}

	return nil
}

func (x *controller) nodeFromTask(t *flow.Task) (*node, error) {
	if x.ranks == nil {
		x.ranks = map[string]int{}
	}

	x.ranks[t.Path().String()] += 1

	n := x.loadOrInit(t.Path())

	for _, d := range t.Dependencies() {
		dep, err := x.nodeFromTask(d)
		if err != nil {
			return nil, err
		}
		n.addDep(dep)
	}

	return n, nil
}

func (x *controller) loadOrInit(p cue.Path) *node {
	if x.nodes == nil {
		x.nodes = map[string]*node{}
	}
	if found, ok := x.nodes[p.String()]; ok {
		return found
	}
	created := &node{path: p, controller: x}
	x.nodes[p.String()] = created
	return created

}

type Node interface {
	Path() cue.Path
	Deps() []cue.Path
}

type node struct {
	*controller

	path cue.Path
	deps []*node
}

func (n *node) String() string {
	return n.path.String()
}

func (n *node) Path() cue.Path {
	return n.path
}

func (n *node) Deps() []cue.Path {
	deps := make([]cue.Path, len(n.deps))
	for i := range n.deps {
		deps[i] = n.deps[i].path
	}
	return deps
}

func (n *node) addDep(dep *node) {
	added := false
	for _, d := range n.deps {
		if d == dep {
			added = true
		}
	}

	if !added {
		n.deps = append(n.deps, dep)
	}
}

func (n *node) cycleCheck() bool {
	for d := range n.rangeDep() {
		if n == d {
			return true
		}
	}

	return false
}

func (n *node) rangeDep() func(yield func(n *node) bool) {
	return func(yield func(n *node) bool) {
		for _, d := range n.deps {
			if !yield(d) {
				return
			}

			for s := range d.rangeDep() {
				if !yield(s) {
					return
				}
			}
		}
	}

}

func (n *node) Run(ctx context.Context) error {
	if n.cycleCheck() {
		return errors.Errorf("cycle deps: %s", n.path)
	}

	if len(n.deps) > 0 {
		errOnce := NewErrOnce()

		for _, d := range n.deps {
			errOnce.Go(func() error {
				return d.Run(ctx)
			})
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-errOnce.Done():
			if err := errOnce.Err(); err != nil {
				return err
			}
		}
	}

	return n.runOnce(ctx, n)
}

func NewErrOnce() *ErrOnce {
	return &ErrOnce{
		d: make(chan struct{}),
	}
}

type ErrOnce struct {
	wg   sync.WaitGroup
	once sync.Once
	d    chan struct{}
	err  error
}

func (e *ErrOnce) done(err error) {
	e.once.Do(func() {
		e.err = err
		e.d <- struct{}{}
	})
}

func (e *ErrOnce) Go(do func() error) {
	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		if err := do(); err != nil {
			e.done(err)
		}
	}()
}

func (e *ErrOnce) Err() error {
	return e.err
}

func (e *ErrOnce) Done() <-chan struct{} {
	go func() {
		e.wg.Wait()
		e.done(nil)
	}()

	return e.d
}
