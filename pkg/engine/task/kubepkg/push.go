package kubepkg

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/octohelm/crkit/pkg/artifact/kubepkg"
	"github.com/octohelm/crkit/pkg/oci/remote"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	taskocitar "github.com/octohelm/piper/pkg/engine/task/ocitar"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Push{})
}

type Push struct {
	task.Task

	// KubePkg spec
	KubePkg KubePkg `json:"kubepkg"`

	// Platforms of oci tar, if empty it will based on KubePkg
	Platforms []string `json:"platforms,omitzero"`

	// WithAnnotations pick annotations of KubePkg as image annotations
	WithAnnotations []string `json:"withAnnotations,omitzero"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename taskocitar.Rename `json:"rename,omitzero"`

	// HostAliases to switch registry target
	HostAliases map[string]string `json:"hostAliases,omitzero"`

	// RemoteURL deprecated use HostAliases instead
	RemoteURL string `json:"remoteURL,omitzero"`
}

func (t *Push) Do(ctx context.Context) error {
	wd, err := enginetask.ClientContext.From(ctx).SourceDir(ctx)
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

	if t.RemoteURL != "" {
		u, err := url.Parse(t.RemoteURL)
		if err == nil {
			if t.HostAliases == nil {
				t.HostAliases = map[string]string{}
			}
			t.HostAliases["docker.io"] = u.Host
		}
	}

	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)
	ns, err := container.NewNamespace(ctx, registryAuthStore, container.NamespaceOptions{
		CacheDir: cacheDir,
	})
	if err != nil {
		return err
	}

	packer := &kubepkg.Packer{
		Namespace:       ns,
		Renamer:         t.Rename.Renamer,
		WithAnnotations: t.WithAnnotations,
		Platforms:       t.Platforms,
	}

	kpkg := kubepkgv1alpha1.KubePkg(t.KubePkg)

	idx, err := packer.PackAsIndex(ctx, &kpkg)
	if err != nil {
		return fmt.Errorf("pack failed: %w", err)
	}

	ns2, err := container.NewNamespace(ctx, registryAuthStore, container.NamespaceOptions{
		HostAliases: t.HostAliases,
	})
	if err != nil {
		return err
	}

	return remote.PushIndex(ctx, idx, ns2)
}
