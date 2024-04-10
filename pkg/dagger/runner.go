package dagger

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sync"

	"github.com/dagger/dagger/telemetry/sdklog"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/dagger/dagger/engine"
	"github.com/dagger/dagger/engine/client"
	contextx "github.com/octohelm/x/context"
	"golang.org/x/sync/errgroup"
)

func WithLogExporter(exporter sdklog.LogExporter) EngineOptionFunc {
	return func(x *options) {
		x.EngineLogs = exporter
	}
}

func WithSpanExporter(exporter trace.SpanExporter) EngineOptionFunc {
	return func(x *options) {
		x.EngineTrace = exporter
	}
}

type EngineOptionFunc = func(x *options)

type options struct {
	client.Params

	Hosts
}

func (o *options) Build(optFns ...EngineOptionFunc) {
	for i := range optFns {
		optFns[i](o)
	}

	if o.Hosts.Default == nil {
		o.AddHost(RunnerHost())
	}
}

var engineVersion = sync.OnceValue(func() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range bi.Deps {
			if dep.Path == "github.com/dagger/dagger" {
				return dep.Version
			}
		}
	}
	return "v0.11.0"
})

func init() {
	// ugly to set engine.Version
	engine.Version = engineVersion()
}

var DefaultRunnerHost = fmt.Sprintf("docker-image://ghcr.io/dagger/engine:%s", engineVersion())

func RunnerHost() *PiperRunnerHost {
	return &PiperRunnerHost{
		Name:       DefaultRunnerHost,
		RunnerHost: DefaultRunnerHost,
	}
}

var RunnerContext = contextx.New[Runner]()

func NewRunner(optFns ...EngineOptionFunc) (Runner, error) {
	opt := &options{}
	opt.Build(optFns...)

	if p := os.Getenv("PIPER_BUILDER_HOST"); p != "" {
		runnerHosts, err := ParsePiperRunnerHosts(p)
		if err != nil {
			return nil, err
		}
		for _, x := range runnerHosts {
			opt.AddHost(&x)
		}
	}

	return &runner{options: opt}, nil
}

type Runner interface {
	Select(ctx context.Context, scope Scope) Engine
	Shutdown(ctx context.Context) error
}

type runner struct {
	*options

	engines sync.Map
}

func (r *runner) ClientParams(scope *Scope) client.Params {
	runnerHost := r.GetHost(scope.Platform)
	p := r.Params
	p.RunnerHost = runnerHost.RunnerHost
	scope.ID = runnerHost.Name
	return p
}

func (r *runner) Select(ctx context.Context, scope Scope) Engine {
	if scope.Platform == "" {
		scope.Platform = Platform(fmt.Sprintf("linux/%s", runtime.GOARCH))
	}

	if v, ok := r.engines.Load(scope.ID); ok {
		return v.(Engine)
	}

	e := NewEngine(scope, r.ClientParams(&scope))
	r.engines.Store(scope.ID, e)
	return e
}

func (r *runner) Shutdown(ctx context.Context) error {
	eg := &errgroup.Group{}

	r.engines.Range(func(key, value any) bool {
		e := value.(Engine)
		eg.Go(func() error {
			return e.Shutdown(ctx)
		})
		return true
	})

	return eg.Wait()
}

type Hosts struct {
	Default    *PiperRunnerHost
	Platformed map[Platform][]*PiperRunnerHost
}

func (h *Hosts) GetHost(platform Platform) *PiperRunnerHost {
	if h.Platformed != nil {
		platformed, ok := h.Platformed[platform]
		if ok {
			return platformed[rand.Intn(len(platformed))]
		}
	}
	return h.Default
}

func (h *Hosts) AddHost(runnerHost *PiperRunnerHost) {
	if len(runnerHost.Platforms) == 0 {
		h.Default = runnerHost
		return
	}

	if h.Platformed == nil {
		h.Platformed = map[Platform][]*PiperRunnerHost{}
	}

	for _, platform := range runnerHost.Platforms {
		p := Platform(platform.String())
		h.Platformed[p] = append(h.Platformed[p], runnerHost)
	}
}

func Select(ctx context.Context, scope Scope) Engine {
	return RunnerContext.From(ctx).Select(ctx, scope)
}
