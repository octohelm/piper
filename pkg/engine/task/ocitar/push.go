package ocitar

import (
	"context"
	"io"
	"os"

	"github.com/octohelm/crkit/pkg/oci/remote"
	ocitar "github.com/octohelm/crkit/pkg/oci/tar"
	"github.com/octohelm/cuekit/pkg/cueflow/task"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Push{})
}

type Push struct {
	task.Task

	// SrcFile of oci tar
	SrcFile file.File `json:"srcFile"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename Rename `json:"rename,omitzero"`

	// HostAliases to switch registry target
	HostAliases map[string]string `json:"hostAliases,omitzero"`
}

func (t *Push) Do(ctx context.Context) error {
	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)

	ns, err := container.NewNamespace(ctx, registryAuthStore, container.NamespaceOptions{
		HostAliases: t.HostAliases,
	})
	if err != nil {
		return err
	}

	return t.SrcFile.WorkDir.Do(ctx, func(ctx context.Context, wd pkgwd.WorkDir) error {
		idx, err := ocitar.Index(func() (io.ReadCloser, error) {
			return wd.OpenFile(ctx, t.SrcFile.Filename, os.O_RDONLY, os.ModePerm)
		})
		if err != nil {
			return err
		}

		return remote.PushIndex(ctx, idx, ns)
	})
}
