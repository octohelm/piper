package engine

import (
	"os"
	"path"
	"strings"

	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	cueload "cuelang.org/go/cue/load"
	"github.com/octohelm/cuemod/pkg/cuemod"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/octohelm/piper/cuepkg"
	"github.com/octohelm/piper/pkg/engine/plan"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"

	_ "github.com/octohelm/piper/pkg/engine/plan/task"
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
	entryFile       string
	output          string
	imagePullPrefix string
}

type OptFunc = func(o *option)

func WithPlan(root string) OptFunc {
	return func(o *option) {
		o.entryFile = root
	}
}

func WithOutput(output string) OptFunc {
	return func(o *option) {
		o.output = output
	}
}

var inCI = os.Getenv("CI") == "true"

func New(ctx context.Context, opts ...OptFunc) (Project, error) {
	c := &project{}
	for i := range opts {
		opts[i](&c.opt)
	}

	cwd, _ := os.Getwd()
	sourceRoot := path.Join(cwd, c.opt.entryFile)

	if strings.HasSuffix(sourceRoot, ".cue") {
		sourceRoot = path.Dir(sourceRoot)
	}

	c.sourceRoot = sourceRoot

	buildConfig := cuemod.ContextFor(cwd).BuildConfig(ctx)

	instances := cueload.Instances([]string{c.opt.entryFile}, buildConfig)
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
	opt        option
	sourceRoot string
	instance   *build.Instance
}

func (p *project) PlanRoot() string {
	return p.sourceRoot
}

func (p *project) Run(ctx context.Context, action ...string) error {
	val := cuecontext.New().BuildInstance(p.instance)
	if err := val.Err(); err != nil {
		return err
	}

	cueValue := plan.WrapValue(val)

	runner := plan.NewRunner(cueValue, p.opt.output)

	ctx = plan.TaskRunnerFactoryContext.Inject(ctx, core.DefaultFactory)
	ctx = plan.ContextContext.Inject(ctx, p)

	return runner.Run(ctx, action)
}
