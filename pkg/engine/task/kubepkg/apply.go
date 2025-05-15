package kubepkg

import (
	"context"
	"log/slog"

	"github.com/go-courier/logr"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/kubekit/pkg/kubeclient"
	"github.com/octohelm/kubepkgspec/pkg/object"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	taskclient "github.com/octohelm/piper/pkg/engine/task/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	enginetask.Registry.Register(&Apply{})
}

// Apply to kubernetes
type Apply struct {
	task.Task

	// Kubeconfig path
	Kubeconfig string `json:"kubeconfig"`

	// Manifests of k8s resources
	Manifests []taskclient.Any `json:"manifests"`
}

func (r *Apply) Do(ctx context.Context) error {
	c, err := kubeclient.NewClient(r.Kubeconfig)
	if err != nil {
		return err
	}

	log := logr.FromContext(ctx)

	for _, m := range r.Manifests {
		data, err := m.MarshalJSON()
		if err != nil {
			return err
		}

		u := &unstructured.Unstructured{}
		if err := u.UnmarshalJSON(data); err != nil {
			return err
		}

		o, err := object.FromRuntimeObject(u)
		if err != nil {
			return err
		}

		gvk := o.GetObjectKind().GroupVersionKind()

		log.WithValues(
			slog.String("name", o.GetName()),
			slog.String("namespace", o.GetNamespace()),
			slog.String("gvk", gvk.String()),
		).Info("Applying")

		if err := c.Patch(ctx, o, client.Apply, client.FieldOwner("kubepkg"), client.ForceOwnership); err != nil {
			return err
		}
	}

	return nil
}
