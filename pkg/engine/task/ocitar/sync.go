package ocitar

import (
	"context"

	"github.com/octohelm/crkit/pkg/artifact/kubepkg"
	contentremote "github.com/octohelm/crkit/pkg/content/remote"
	"github.com/octohelm/crkit/pkg/oci/remote"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/workload"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/container"
)

func init() {
	enginetask.Registry.Register(&Sync{})
}

type Sync struct {
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
}

func (t *Sync) Do(ctx context.Context) error {
	registryAuthStore := container.RegistryAuthStoreContext.From(ctx)
	registryHosts := contentremote.RegistryHosts{}

	for auth := range registryAuthStore.RegistryAuths(ctx) {
		registryHosts[auth.Host] = auth.RegistryHost
	}

	ns, err := contentremote.New(ctx, registryHosts)
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

	kpkg := &kubepkgv1alpha1.KubePkg{}
	kpkg.Name = "oci-tar"
	kpkg.Annotations = t.Annotations
	kpkg.Spec.Version = "v0.0.0"
	kpkg.Spec.Images = make(map[string]kubepkgv1alpha1.Image)
	kpkg.Spec.Images["x"] = *workload.ParseImage(t.Source)

	idx, err := packer.PackAsIndex(ctx, kpkg)
	if err != nil {
		return err
	}

	return remote.PushIndex(ctx, idx, ns)
}
