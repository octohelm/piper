package internal

import (
	"context"
	"cuelang.org/go/cue"
	"fmt"
	"github.com/pkg/errors"
)

type Node interface {
	fmt.Stringer

	Path() cue.Path
	Deps() []Node
}

type node struct {
	*Controller

	path cue.Path
	deps []*node
}

func (n *node) String() string {
	return FormatAsJSONPath(n.path)
}

func (n *node) Path() cue.Path {
	return n.path
}

func (n *node) Deps() []Node {
	deps := make([]Node, len(n.deps))
	for i := range n.deps {
		deps[i] = n.deps[i]
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
		if ps, ok := PlatformScoped(n.path); ok {
			if depPs, ok := PlatformScoped(dep.path); ok {
				if !ps.Equals(*depPs) {
					return
				}
			}
		}

		n.deps = append(n.deps, dep)
	}
}

func (n *node) cycledDeps() bool {
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
	if n.cycledDeps() {
		deps := []*node{n}

		for d := range n.rangeDep() {
			deps = append(deps, d)
			if n == d {
				break
			}

		}

		return errors.Errorf("cycle deps: %s", deps)
	}

	if len(n.deps) > 0 {
		eg, c := WithContext(ctx)

		for _, d := range n.deps {
			eg.Go(func() error {
				return d.Run(c)
			})
		}

		if err := eg.Wait(); err != nil {
			// ignore cancel
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
	}

	return n.runOnce(ctx, n)
}
