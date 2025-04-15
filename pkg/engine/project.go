package engine

import (
	"context"
	"iter"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/interpreter/embed"
	"cuelang.org/go/cue/interpreter/wasm"
	"github.com/k0sproject/rig"
	cuekitcuecontext "github.com/octohelm/cuekit/pkg/cuecontext"
	"github.com/octohelm/cuekit/pkg/mod/module"
	"github.com/octohelm/piper/cuepkg"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"

	_ "github.com/octohelm/piper/pkg/engine/task/archive"
	_ "github.com/octohelm/piper/pkg/engine/task/client"
	_ "github.com/octohelm/piper/pkg/engine/task/container"
	_ "github.com/octohelm/piper/pkg/engine/task/encoding"
	_ "github.com/octohelm/piper/pkg/engine/task/exec"
	_ "github.com/octohelm/piper/pkg/engine/task/file"
	_ "github.com/octohelm/piper/pkg/engine/task/http"
	_ "github.com/octohelm/piper/pkg/engine/task/kubepkg"
	_ "github.com/octohelm/piper/pkg/engine/task/wd"
)

func init() {
	if err := cuepkg.RegisterAsMemModule(); err != nil {
		panic(err)
	}
}

type Project interface {
	Run(ctx context.Context, action ...string) error
}

type option struct {
	cwd      string
	entry    string
	cacheDir string
}

type OptFunc = func(o *option)

func WithProject(root string) OptFunc {
	return func(o *option) {
		o.entry = root
	}
}

func WithCacheDir(root string) OptFunc {
	return func(o *option) {
		o.cacheDir = root
	}
}

func New(ctx context.Context, opts ...OptFunc) (Project, error) {
	c := &project{}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	c.opt.cacheDir = filepath.Join(cacheDir, "piper")

	for i := range opts {
		opts[i](&c.opt)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	c.opt.cwd = cwd

	c.client, err = fromOpt(cwd, &c.opt)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type project struct {
	opt option

	*client
}

func seq[T any](values ...T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

func (p *project) Run(ctx context.Context, action ...string) error {
	runner := cueflow.NewRunner(func() (cueflow.Value, *module.Module, error) {
		buildConfig, err := cuekitcuecontext.NewConfig(cuekitcuecontext.WithRoot(p.opt.cwd))
		if err != nil {
			return nil, nil, err
		}

		val, err := cuekitcuecontext.Build(
			buildConfig.Config,
			seq(p.opt.entry),
			cuecontext.Interpreter(embed.New()),
			cuecontext.Interpreter(wasm.New()),
			// cuecontext.EvaluatorVersion(cuecontext.EvalV3),
		)
		if err != nil {
			return nil, nil, err
		}

		return cueflow.WrapValue(val), buildConfig.Module, nil
	})

	ctx = cueflow.TaskRunnerFactoryContext.Inject(ctx, task.Factory)
	ctx = task.ClientContext.Inject(ctx, p)

	return runner.Run(ctx, action)
}

func fromOpt(cwd string, opt *option) (*client, error) {
	return &client{
		cacheDir: sync.OnceValues(func() (wd.WorkDir, error) {
			cacheDir := opt.cacheDir

			if strings.HasPrefix(cacheDir, "~") {
				cacheDir = path.Join(os.Getenv("HOME"), "."+cacheDir[1:])
			}

			return wd.Wrap(
				&rig.Connection{
					Localhost: &rig.Localhost{
						Enabled: true,
					},
				},
				wd.WithDir(cacheDir),
			)
		}),
		sourceDir: sync.OnceValues(func() (wd.WorkDir, error) {
			root := path.Join(cwd, opt.entry)

			if strings.HasSuffix(root, ".cue") {
				root = path.Dir(root)
			}

			return wd.Wrap(
				&rig.Connection{
					Localhost: &rig.Localhost{
						Enabled: true,
					},
				},
				wd.WithDir(root),
			)
		}),
	}, nil
}

type client struct {
	sourceDir func() (wd.WorkDir, error)
	cacheDir  func() (wd.WorkDir, error)
}

func (c *client) SourceDir(ctx context.Context) (wd.WorkDir, error) {
	return c.sourceDir()
}

func (c *client) CacheDir(ctx context.Context, typ string) (wd.WorkDir, error) {
	w, err := c.cacheDir()
	if err != nil {
		return nil, err
	}
	return wd.With(w, wd.WithDir(typ))
}
