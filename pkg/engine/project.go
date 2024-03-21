package engine

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/k0sproject/rig"
	"github.com/octohelm/cuekit/pkg/cuecontext"
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
	_ "github.com/octohelm/piper/pkg/engine/task/flow"
	_ "github.com/octohelm/piper/pkg/engine/task/http"
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

func (p *project) Run(ctx context.Context, action ...string) error {
	runner := cueflow.NewRunner(func() (cueflow.Value, error) {
		buildConfig, err := cuecontext.NewConfig(cuecontext.WithRoot(p.opt.cwd))
		if err != nil {
			return nil, err
		}

		val, err := cuecontext.Build(buildConfig, p.opt.entry)
		if err != nil {
			return nil, err
		}

		return cueflow.WrapValue(val), nil
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
