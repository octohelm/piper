package engine

import (
	"os"
	"path"
	"strings"
	"sync"

	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
	"github.com/k0sproject/rig"
	"github.com/octohelm/cuemod/pkg/cuemod"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/octohelm/piper/cuepkg"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"

	_ "github.com/octohelm/piper/pkg/engine/task/archive"
	_ "github.com/octohelm/piper/pkg/engine/task/client"
	_ "github.com/octohelm/piper/pkg/engine/task/exec"
	_ "github.com/octohelm/piper/pkg/engine/task/file"
	_ "github.com/octohelm/piper/pkg/engine/task/flow"
	_ "github.com/octohelm/piper/pkg/engine/task/http"
	_ "github.com/octohelm/piper/pkg/engine/task/wd"
)

func init() {
	if err := cuepkg.RegistryCueStdlibs(); err != nil {
		panic(err)
	}
}

type Project interface {
	Run(ctx context.Context, action ...string) error
}

type option struct {
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
	c.opt.cacheDir = "~/.piper/cache"

	for i := range opts {
		opts[i](&c.opt)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	c.client, err = fromOpt(cwd, &c.opt)
	if err != nil {
		return nil, err
	}

	buildConfig := cuemod.ContextFor(cwd).BuildConfig(ctx)

	instances := cueload.Instances([]string{c.opt.entry}, buildConfig)
	if len(instances) != 1 {
		return nil, errors.New("only one package is supported at a time")
	}
	c.instance = instances[0]

	if err := c.instance.Err; err != nil {
		return nil, err
	}

	return c, nil
}

type project struct {
	opt option

	*client

	instance *build.Instance
}

func (p *project) Run(ctx context.Context, action ...string) error {
	val := cuecontext.New().BuildInstance(p.instance)
	if err := val.Err(); err != nil {
		return err
	}

	runner := cueflow.NewRunner(cueflow.WrapValue(val))

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
