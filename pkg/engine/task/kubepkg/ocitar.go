package kubepkg

import (
	"context"
	"encoding"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-courier/logr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/octohelm/crkit/pkg/kubepkg"
	"github.com/octohelm/crkit/pkg/ocitar"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	"github.com/octohelm/piper/pkg/engine/task/file"
	"github.com/octohelm/piper/pkg/otel"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &OciTar{})
}

type OciTar struct {
	task.Task

	// KubePkg spec
	KubePkg KubePkg `json:"kubepkg"`

	// Platforms of oci tar, if empty it will based on KubePkg
	Platforms []string `json:"platforms,omitempty"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename Rename `json:"rename,omitempty"`
	// OutFile of OciTar
	OutFile file.File `json:"outFile"`

	File file.File
}

var _ encoding.TextUnmarshaler = &Rename{}

type Rename struct {
	renamer kubepkg.Renamer
}

func (r *Rename) CueType() []byte {
	return []byte("string")
}

func (r *Rename) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	renamer, err := kubepkg.NewTemplateRenamer(string(text))
	if err != nil {
		return err
	}

	*r = Rename{
		renamer: renamer,
	}

	return nil
}

func (t *OciTar) Do(ctx context.Context) error {
	wd, err := task.ClientContext.From(ctx).SourceDir(ctx)
	if err != nil {
		return err
	}

	cacheWd, err := pkgwd.With(wd, pkgwd.WithDir(filepath.Join(".piper", "cache", "blobs")))
	if err != nil {
		return err
	}

	cacheDir, err := pkgwd.RealPath(cacheWd)
	if err != nil {
		return err
	}

	l := logr.FromContext(ctx)

	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)

	process := make(chan ocitar.Update)
	defer func() {
		close(process)
	}()

	loggerGetters := &sync.Map{}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	getLogger := func(d string) func() logr.Logger {
		value, _ := loggerGetters.LoadOrStore(d, sync.OnceValue(func() logr.Logger {
			_, log := l.Start(ctx, d)
			return log
		}))

		return value.(func() logr.Logger)
	}

	lastUpdate := atomic.Pointer[ocitar.Update]{}

	go func() {
		for range tick.C {
			if u := lastUpdate.Load(); u != nil {
				getLogger(u.String())().WithValues(
					slog.Int64(otel.LogAttrProgressTotal, u.Total),
					slog.Int64(otel.LogAttrProgressCurrent, u.Complete),
				).Info("pulling")
			}
		}
	}()

	go func() {
		for u := range process {
			lastUpdate.Store(&u)

			if u.Complete == u.Total {
				getLogger(u.String())().WithValues(
					slog.Int64(otel.LogAttrProgressTotal, u.Total),
					slog.Int64(otel.LogAttrProgressCurrent, u.Complete),
				).Info("pulling")
			}
		}
	}()

	packer := &kubepkg.Packer{
		Cache: cache.NewFilesystemCache(cacheDir),
		CreatePuller: func(repo name.Repository, options ...remote.Option) (*remote.Puller, error) {
			for auth := range registryAuthStore.RegistryAuths(ctx) {
				if repo.RegistryStr() == auth.Address {
					options = append(options, remote.WithAuth(authn.FromConfig(authn.AuthConfig{
						Username: auth.Username,
						Password: auth.Password,
					})))
				}
			}

			options = append(options)

			return remote.NewPuller(options...)
		},
		Renamer:   t.Rename.renamer,
		Platforms: t.Platforms,
	}

	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) error {
		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "open file failed")
		}
		defer f.Close()

		kpkg := kubepkgv1alpha1.KubePkg(t.KubePkg)

		idx, err := packer.PackAsIndex(ctx, &kpkg)
		if err != nil {
			return err
		}

		if err := ocitar.Write(f, idx, ocitar.WithProgress(process)); err != nil {
			return errors.Errorf("%#v", err)
		}

		return nil
	})
}
