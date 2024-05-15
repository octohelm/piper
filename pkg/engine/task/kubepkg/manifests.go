package kubepkg

import (
	"context"

	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/kubepkg"
	"github.com/octohelm/kubepkgspec/pkg/manifest"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
	"github.com/octohelm/x/anyjson"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Manifests{})
}

// Manifests extract manifests from KubePkg
type Manifests struct {
	task.Task

	// KubePkg spec
	KubePkg KubePkg `json:"kubepkg"`
	// recursively extract KubePkg in sub manifests
	Recursive bool `json:"recursive,omitempty"`
	// Manifests of k8s resources
	Manifests []client.Any `json:"-" output:"manifests"`
}

func (r *Manifests) Do(ctx context.Context) error {
	kpkg := kubepkgv1alpha1.KubePkg(r.KubePkg)

	manifests, err := manifest.SortedExtract(&kpkg, kubepkg.WithRecursive(r.Recursive))
	if err != nil {
		return err
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
