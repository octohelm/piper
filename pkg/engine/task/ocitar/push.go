package ocitar

import (
	"context"
	"io"
	"os"

	"github.com/octohelm/piper/internal/pkg/processpool"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/octohelm/crkit/pkg/kubepkg"
	"github.com/octohelm/crkit/pkg/ocitar"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Push{})
}

type Push struct {
	task.Task

	// SrcFile of oci tar
	SrcFile file.File `json:"srcFile"`

	// RemoteURL of container registry
	RemoteURL string `json:"remoteURL"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename Rename `json:"rename,omitempty"`
}

func (t *Push) registry() (kubepkg.Registry, error) {
	if t.RemoteURL != "" {
		return kubepkg.NewRegistry(t.RemoteURL)
	}
	return nil, nil
}

func (t *Push) Do(ctx context.Context) error {
	r, err := t.registry()
	if err != nil {
		return err
	}

	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)
	p := processpool.NewProcessPool("pushing")
	go p.Wait(ctx)
	defer func() {
		_ = p.Close()
	}()

	pusher := &kubepkg.Pusher{
		Registry: r,
		Renamer:  t.Rename.Renamer,
		CreatePusher: func(ref name.Reference, options ...remote.Option) (*remote.Pusher, error) {
			for auth := range registryAuthStore.RegistryAuths(ctx) {
				if ref.Context().RegistryStr() == auth.Address {
					options = append(options, remote.WithAuth(authn.FromConfig(authn.AuthConfig{
						Username: auth.Username,
						Password: auth.Password,
					})))
				}
			}
			return remote.NewPusher(append(options, remote.WithContext(ctx), remote.WithProgress(p.Progress(ref)))...)
		},
	}

	return t.SrcFile.WorkDir.Do(ctx, func(ctx context.Context, wd pkgwd.WorkDir) error {
		idx, err := ocitar.Index(func() (io.ReadCloser, error) {
			return wd.OpenFile(ctx, t.SrcFile.Filename, os.O_RDONLY, os.ModePerm)
		})
		if err != nil {
			return err
		}

		return pusher.PushIndex(ctx, idx)
	})
}
