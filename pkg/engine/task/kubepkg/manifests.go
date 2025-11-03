package kubepkg

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/kubepkg"
	"github.com/octohelm/kubepkgspec/pkg/manifest"
	"github.com/octohelm/kubepkgspec/pkg/object"
	"github.com/octohelm/kubepkgspec/pkg/workload"
	"github.com/octohelm/x/anyjson"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	taskocitar "github.com/octohelm/piper/pkg/engine/task/ocitar"
)

func init() {
	enginetask.Registry.Register(&Manifests{})
}

// Manifests extract manifests from KubePkg
type Manifests struct {
	task.Task

	// KubePkg spec
	KubePkg KubePkg `json:"kubepkg"`
	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename taskocitar.Rename `json:"rename,omitzero"`
	// recursively extract KubePkg in sub manifests
	Recursive bool `json:"recursive,omitzero"`
	// Manifests of k8s resources
	Manifests []client.Any `json:"-" output:"manifests"`
}

func (r *Manifests) Do(ctx context.Context) error {
	kpkg := kubepkgv1alpha1.KubePkg(r.KubePkg)

	manifests, err := manifest.SortedExtract(&kpkg, kubepkg.WithRecursive(r.Recursive))
	if err != nil {
		return fmt.Errorf("extract manifests failed: %w", err)
	}

	if renamer := r.Rename.Renamer; renamer != nil {
		images := workload.Images(func(yield func(object.Object) bool) {
			for _, m := range manifests {
				if !yield(m) {
					return
				}
			}
		})

		for img := range images {
			repo, err := name.NewRepository(img.Name)
			if err != nil {
				return err
			}
			img.Name = renamer.Rename(repo)
		}
	}

	r.Manifests = make([]client.Any, len(manifests))

	for i := range manifests {
		v, err := anyjson.FromValue(manifests[i])
		if err != nil {
			return err
		}

		cleaned := anyjson.Merge(anyjson.Valuer(&anyjson.Object{}), v, anyjson.WithEmptyObjectAsNull())

		r.Manifests[i] = client.Any{
			Value: cleaned,
		}
	}

	return nil
}
