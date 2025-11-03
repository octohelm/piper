package kubepkg

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/octohelm/crkit/pkg/artifact/kubepkg"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"

	"github.com/octohelm/piper/internal/pkg/processpool"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
	taskocitar "github.com/octohelm/piper/pkg/engine/task/ocitar"
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

	// RemoteURL of container registry
	RemoteURL string `json:"remoteURL"`
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

	fetcher := &http.Client{
		Transport: remote.DefaultTransport,
	}

	p := processpool.NewProcessPool("pulling")
	go p.Wait(ctx)
	defer func() {
		_ = p.Close()
	}()

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

	pp := processpool.NewProcessPool("pushing")
	go pp.Wait(ctx)
	defer func() {
		_ = pp.Close()
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
			return remote.NewPusher(append(options, remote.WithContext(ctx), remote.WithProgress(pp.Progress(ref)))...)
		},
	}

	kpkg := kubepkgv1alpha1.KubePkg(t.KubePkg)

	idx, err := packer.PackAsIndex(ctx, &kpkg)
	if err != nil {
		return err
	}

	return pusher.PushIndex(ctx, idx)
}
