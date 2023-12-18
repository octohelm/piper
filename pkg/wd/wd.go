package wd

import (
	"context"
	"fmt"
	"github.com/octohelm/unifs/pkg/filesystem/local"
	"strings"

	"github.com/go-courier/logr"
	"github.com/k0sproject/rig"
	"github.com/k0sproject/rig/exec"
	"github.com/octohelm/unifs/pkg/filesystem"
)

type WorkDir interface {
	fmt.Stringer

	filesystem.FileSystem

	Options() Options
	Exec(ctx context.Context, cmd string, optFns ...OptionFunc) error
	ExecOutput(ctx context.Context, cmd string, optFns ...OptionFunc) (string, error)
}

type WorkDirConnection interface {
	Connection() *rig.Connection
}

type OptionFunc func(opt *Options)

type Options struct {
	BasePath Dir
	Env      map[string]string
	User     string
}

func (o *Options) Build(optionFuncs ...OptionFunc) {
	for _, fn := range optionFuncs {
		fn(o)
	}
}

func WithDir(dir string) OptionFunc {
	return func(opt *Options) {
		opt.BasePath = opt.BasePath.With(dir)
	}
}

func WithUser(user string) OptionFunc {
	return func(opt *Options) {
		if user != "" {
			opt.User = user
		}
	}
}

func WithEnv(env map[string]string) OptionFunc {
	return func(opt *Options) {
		opt.Env = env
	}
}

func Wrap(c *rig.Connection, optionFuncs ...OptionFunc) (WorkDir, error) {
	w := &wd{}
	w.connection = c
	w.opt.User = "root"
	w.opt.BasePath = "/"

	w.opt.Build(optionFuncs...)

	if err := w.init(); err != nil {
		return nil, err
	}

	return w, nil
}

func With(source WorkDir, optionFuncs ...OptionFunc) (WorkDir, error) {
	if len(optionFuncs) == 0 {
		return source, nil
	}

	w := &wd{}
	w.connection = source.(WorkDirConnection).Connection()
	w.opt = source.Options()

	w.opt.Build(optionFuncs...)

	if err := w.init(); err != nil {
		return nil, err
	}

	return w, nil
}

type wd struct {
	opt Options
	filesystem.FileSystem
	connection *rig.Connection
}

func (w *wd) Connection() *rig.Connection {
	return w.connection
}

func (w *wd) init() error {
	if !w.connection.IsConnected() {
		if err := w.connection.Connect(); err != nil {
			return err
		}
	}

	if w.connection.Localhost != nil {
		w.FileSystem = local.NewLocalFS(w.opt.BasePath.String())
	} else {
		if w.opt.User == "root" {
			w.FileSystem = filesystem.Sub(WrapRigFS(w.connection.SudoFsys()), w.opt.BasePath.String())
		} else {
			w.FileSystem = filesystem.Sub(WrapRigFS(w.connection.Fsys()), w.opt.BasePath.String())
		}
	}
	return nil
}

func (w *wd) Options() Options {
	return w.opt
}

func (w *wd) String() string {
	switch w.connection.Protocol() {
	case "Local":
		return fmt.Sprintf("%s (%s)", w.opt.BasePath, w.opt.User)
	default:
		return fmt.Sprintf("%s (%s,%s@%s)", w.opt.BasePath, strings.ToLower(w.connection.Protocol()), w.opt.User, w.connection.Address())
	}
}

func (w *wd) Exec(ctx context.Context, cmd string, optFns ...OptionFunc) error {
	logr.FromContext(ctx).WithValues("name", w).Debug(cmd)
	b, opts := w.normalizeExecArgs(cmd, optFns...)
	return w.connection.Exec(b.String(), opts...)
}

func (w *wd) ExecOutput(ctx context.Context, cmd string, optFns ...OptionFunc) (output string, err error) {
	logr.FromContext(ctx).WithValues("name", w).Debug(cmd)
	b, opts := w.normalizeExecArgs(cmd, optFns...)
	return w.connection.ExecOutput(b.String(), opts...)
}

func (w *wd) normalizeExecArgs(cmd string, optFns ...OptionFunc) (b *strings.Builder, execOptions []exec.Option) {
	b = &strings.Builder{}

	opt := &Options{
		BasePath: w.opt.BasePath,
		User:     w.opt.User,
	}

	for _, optFn := range optFns {
		optFn(opt)
	}

	if w.connection.Localhost == nil && opt.User == "root" {
		execOptions = append(execOptions, exec.Sudo(w.connection))
	}

	if opt.BasePath != "" && opt.BasePath != "/" {
		_, _ = fmt.Fprintf(b, "cd %s; ", opt.BasePath)
	}

	for k, v := range opt.Env {
		_, _ = fmt.Fprintf(b, "%s=%s ", k, v)
	}

	b.WriteString(cmd)

	return
}
