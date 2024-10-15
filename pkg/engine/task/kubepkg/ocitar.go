package kubepkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/octohelm/piper/internal/pkg/processpool"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/octohelm/crkit/pkg/kubepkg"
	"github.com/octohelm/crkit/pkg/kubepkg/cache"
	"github.com/octohelm/crkit/pkg/ocitar"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	"github.com/octohelm/piper/pkg/engine/task/file"
	taskocitar "github.com/octohelm/piper/pkg/engine/task/ocitar"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
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

	// WithAnnotations pick annotations of KubePkg as image annotations
	WithAnnotations []string `json:"withAnnotations,omitempty"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename taskocitar.Rename `json:"rename,omitempty"`

	// OutFile of OciTar
	OutFile file.File `json:"outFile"`

	// File of tar created
	File file.File
}

func (t *OciTar) Do(ctx context.Context) error {
	wd, err := task.ClientContext.From(ctx).SourceDir(ctx)
	if err != nil {
		return err
	}

	cacheWd, err := pkgwd.With(wd, pkgwd.WithDir(filepath.Join(".piper", "cache", "registry")))
	if err != nil {
		return err
	}

	cacheDir, err := pkgwd.RealPath(cacheWd)
	if err != nil {
		return err
	}

	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)
	p := processpool.NewProcessPool("pulling")
	go p.Wait(ctx)
	defer func() {
		_ = p.Close()
	}()

	fetcher := &http.Client{
		Transport: remote.DefaultTransport,
	}

	transport := taskocitar.WithRoundTripperFunc(func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/blobs/") {
			resp, err := fetcher.Do(req)
			if err != nil {
				return nil, err
			}

			if resp.ContentLength > 0 {
				r, _ := name.NewRegistry(req.Host)
				parts := strings.Split(strings.Split(req.URL.Path, "/v2/")[1], "/blobs/")
				ref := r.Repo(parts[0]).Digest(parts[1])

				resp.Body = processpool.NewProcessReader(resp.Body, resp.ContentLength, p.Progress(ref))
			}

			return resp, nil
		}

		return next.RoundTrip(req)
	})(remote.DefaultTransport)

	packer := &kubepkg.Packer{
		Cache:           cache.NewFilesystemCache(cacheDir),
		Renamer:         t.Rename.Renamer,
		WithAnnotations: t.WithAnnotations,
		Platforms:       t.Platforms,
		CreatePuller: func(ref name.Reference, options ...remote.Option) (*remote.Puller, error) {
			for auth := range registryAuthStore.RegistryAuths(ctx) {
				if ref.Context().RegistryStr() == auth.Address {
					options = append(options, remote.WithAuth(authn.FromConfig(authn.AuthConfig{
						Username: auth.Username,
						Password: auth.Password,
					})))
				}
			}
			return remote.NewPuller(append(
				options,
				remote.WithContext(ctx),
				remote.WithTransport(transport),
			)...)
		},
	}

	return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, cwd pkgwd.WorkDir) error {
		if err := filesystem.MkdirAll(ctx, cwd, path.Dir(t.OutFile.Filename)); err != nil {
			return err
		}

		f, err := cwd.OpenFile(ctx, t.OutFile.Filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return fmt.Errorf("open %s failed: %w", t.OutFile.Filename, err)
		}
		defer f.Close()

		kpkg := kubepkgv1alpha1.KubePkg(t.KubePkg)

		idx, err := packer.PackAsIndex(ctx, &kpkg)
		if err != nil {
			return err
		}

		if err := ocitar.Write(f, idx); err != nil {
			return fmt.Errorf("%#v", err)
		}

		return nil
	})
}
