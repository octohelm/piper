package wd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/octohelm/unifs/pkg/filesystem/local"
	"github.com/opencontainers/go-digest"

	"github.com/go-courier/logr"
	"github.com/k0sproject/rig"
	"github.com/k0sproject/rig/exec"
	"github.com/octohelm/unifs/pkg/filesystem"
)

type WorkDir interface {
	fmt.Stringer

	filesystem.FileSystem

	Options() Options

	Digest(ctx context.Context, path string) (digest.Digest, error)

	Exec(ctx context.Context, cmd string, optFns ...OptionFunc) error

	Addr() *url.URL
}

type WorkDirConnection interface {
	Connection() *rig.Connection
}

type OptionFunc func(opt *Options)

type Options struct {
	BasePath Dir
	Env      map[string]string
	User     string
	Stdout   io.Writer
	Stderr   io.Writer
}

func (o *Options) Build(optionFuncs ...OptionFunc) {
	for _, fn := range optionFuncs {
		fn(o)
	}
}

func (o *Options) Sudo() bool {
	return o.User == "root"
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

func WithStdout(stdout io.Writer) OptionFunc {
	return func(opt *Options) {
		if stdout != nil {
			opt.Stdout = stdout
		}
	}
}

func WithStderr(stderr io.Writer) OptionFunc {
	return func(opt *Options) {
		if stderr != nil {
			opt.Stderr = stderr
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

func (w *wd) Digest(ctx context.Context, filename string) (digest.Digest, error) {
	if w.connection.Localhost != nil {
		f, err := w.OpenFile(ctx, filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return "", nil
		}
		defer f.Close()
		d, err := digest.FromReader(f)
		if err != nil {
			return "", err
		}
		return d, nil
	}
	hex, err := w.connection.SudoFsys().Sha256(w.opt.BasePath.With(filename).String())
	if err != nil {
		return "", err
	}
	return digest.NewDigestFromHex(string(digest.SHA256), hex), nil
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
		if err := os.MkdirAll(w.opt.BasePath.String(), os.ModePerm); err != nil {
			return err
		}
		w.FileSystem = local.NewFS(w.opt.BasePath.String())
	} else if w.connection.SSH != nil && w.opt.Sudo() {
		sc := w.connection.SSH.Client()
		c, err := WrapSFTP(sc)
		if err != nil {
			return err
		}
		w.FileSystem = filesystem.Sub(c, w.opt.BasePath.String())
	} else {
		if w.opt.Sudo() {
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

func (w *wd) Addr() *url.URL {
	switch w.connection.Protocol() {
	case "Local":
		return &url.URL{
			Scheme: "file",
			User:   url.User(w.opt.User),
			Path:   string(w.opt.BasePath),
		}
	default:
		return &url.URL{
			Scheme: strings.ToLower(w.connection.Protocol()),
			User:   url.User(w.opt.User),
			Host:   w.connection.Address(),
			Path:   string(w.opt.BasePath),
		}
	}
}

func (w *wd) String() string {
	return w.Addr().String()
}

func (w *wd) Exec(ctx context.Context, cmd string, optFns ...OptionFunc) error {
	logr.FromContext(ctx).Info(cmd)

	o, b, opts := w.normalizeExecArgs(cmd, optFns...)

	waiter, err := w.connection.ExecStreams(b.String(), nil, o.Stdout, o.Stderr, opts...)
	if err != nil {
		return err
	}
	return waiter.Wait()
}

func (w *wd) normalizeExecArgs(cmd string, optFns ...OptionFunc) (opt *Options, b *strings.Builder, execOptions []exec.Option) {
	b = &strings.Builder{}

	opt = &Options{
		BasePath: w.opt.BasePath,
		User:     w.opt.User,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
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

func RealPath(dir WorkDir) (string, error) {
	switch x := dir.(type) {
	case *wd:
		return x.opt.BasePath.String(), nil
	}
	return "", errors.New("unsupported")
}

func SameFileSystem(a *url.URL, b *url.URL) bool {
	return a.Scheme == b.Scheme && a.Host == b.Host
}
