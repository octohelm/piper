package ocitar

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/octohelm/crkit/pkg/artifact/kubepkg"
	ocitar "github.com/octohelm/crkit/pkg/oci/tar"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/workload"
	"github.com/octohelm/unifs/pkg/filesystem"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Pull{})
}

type Pull struct {
	task.Task

	// image from
	Source string `json:"source"`

	// Platforms of oci tar, if empty it will based on KubePkg
	Platforms []string `json:"platforms,omitzero"`

	// Annotations
	Annotations map[string]string `json:"annotations,omitzero"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename Rename `json:"rename,omitzero"`

	// OutFile of OciTar
	OutFile file.File `json:"outFile"`

	// File of tar created
	File file.File `json:"-" output:"file"`
}

func (t *Pull) Do(ctx context.Context) error {
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
		Platforms:       t.Platforms,
		WithAnnotations: []string{"*"},
		ImageOnly:       true,
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

		kpkg := &kubepkgv1alpha1.KubePkg{}
		kpkg.Name = "oci-tar"
		kpkg.Annotations = t.Annotations
		kpkg.Spec.Version = "v0.0.0"
		kpkg.Spec.Containers = make(map[string]kubepkgv1alpha1.Container)
		kpkg.Spec.Containers["x"] = kubepkgv1alpha1.Container{
			Image: *workload.ParseImage(t.Source),
		}

		idx, err := packer.PackAsIndex(ctx, kpkg)
		if err != nil {
			return err
		}

		if err := ocitar.Write(f, idx); err != nil {
			return err
		}

		return nil
	})
}
