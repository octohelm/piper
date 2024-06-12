package dagger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"

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
	return &engineImpl{params: params}
}

type engineImpl struct {
	scope  Scope
	params Params

	engineClient *engineclient.Client
	err          error

	once sync.Once
	mu   sync.RWMutex
}

func (e *engineImpl) Shutdown(ctx context.Context) error {
	if e.engineClient != nil {
		return e.engineClient.Close()
	}
	return nil
}

func (e *engineImpl) Scope() Scope {
	return e.scope
}

func (e *engineImpl) conn(ctx context.Context) (Conn, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.once.Do(func() {
		e.engineClient, _, e.err = engineclient.Connect(ctx, e.params)
	})

	return EngineConn(e.engineClient), e.err
}

func (e *engineImpl) Do(ctx context.Context, do func(ctx context.Context, client *Client) error) error {
	engineConn, err := e.conn(ctx)
	if err != nil {
		return err
	}

	c, err := dagger.Connect(ctx, dagger.WithConn(engineConn), dagger.WithSkipCompatibilityCheck())
	if err != nil {
		return err
	}
	defer c.Close()

	return do(ScopeContext.Inject(ctx, e.scope), c)
}

func EngineConn(engineClient *engineclient.Client) Conn {
	return func(req *http.Request) (*http.Response, error) {
		req.SetBasicAuth(engineClient.SecretToken, "")
		resp := httptest.NewRecorder()
		engineClient.ServeHTTP(resp, req)
		return resp.Result(), nil
	}
}

type Conn func(*http.Request) (*http.Response, error)

func (f Conn) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

func (f Conn) Host() string {
	return ":mem:"
}

func (f Conn) Close() error {
	return nil
}
