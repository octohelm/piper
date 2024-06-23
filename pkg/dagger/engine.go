package dagger

import (
	"context"
	"sync"

	"github.com/moby/buildkit/identity"

	"dagger.io/dagger"
	engineclient "github.com/dagger/dagger/engine/client"
	contextx "github.com/octohelm/x/context"
)

type (
	Client   = dagger.Client
	Platform = dagger.Platform
)

type Params = engineclient.Params

var ScopeContext = contextx.New[Scope](contextx.WithDefaultsFunc(func() Scope {
	return Scope{}
}))

type Scope struct {
	Platform Platform
	ID       string
}

type EngineWithScope interface {
	Engine
	Scope() Scope
}

type Engine interface {
	Shutdown(ctx context.Context) error
	Do(ctx context.Context, do func(ctx context.Context, client *Client) error) error
}

func NewEngine(params Params) Engine {
	params.ID = identity.NewID()

	return &engineImpl{params: params}
}

type engineImpl struct {
	params Params

	m            sync.Map
	daggerClient *dagger.Client
}

func (e *engineImpl) Shutdown(ctx context.Context) error {
	if e.daggerClient != nil {
		return e.daggerClient.Close()
	}
	return nil
}

func (e *engineImpl) client(ctx context.Context) (*dagger.Client, error) {
	v, _ := e.m.LoadOrStore("", sync.OnceValues(func() (*dagger.Client, error) {
		engineClient, c, err := engineclient.Connect(context.WithoutCancel(ctx), e.params)
		if err != nil {
			return nil, err
		}

		e.daggerClient, err = dagger.Connect(
			c,
			dagger.WithConn(engineclient.EngineConn(engineClient)),
			dagger.WithSkipCompatibilityCheck(),
		)

		return e.daggerClient, err
	}))

	return v.(func() (*dagger.Client, error))()
}

func (e *engineImpl) Do(ctx context.Context, do func(ctx context.Context, client *Client) error) error {
	c, err := e.client(ctx)
	if err != nil {
		return err
	}
	return do(ctx, c)
}
