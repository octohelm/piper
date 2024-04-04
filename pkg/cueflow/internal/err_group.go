package internal

import (
	"context"
	"sync"
)

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

type Group struct {
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func (g *Group) done() {
	g.wg.Done()
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.done()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
