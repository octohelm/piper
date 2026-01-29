package container

import (
	"context"

	"github.com/octohelm/crkit/pkg/content"
	contentfs "github.com/octohelm/crkit/pkg/content/fs"
	contentproxy "github.com/octohelm/crkit/pkg/content/proxy"
	contentremote "github.com/octohelm/crkit/pkg/content/remote"
	"github.com/octohelm/unifs/pkg/filesystem/local"
)

type NamespaceOptions struct {
	HostAliases map[string]string
	CacheDir    string
}

func NewNamespace(ctx context.Context, registryAuthStore RegistryAuthStore, o NamespaceOptions) (content.Namespace, error) {
	registryHosts := contentremote.RegistryHosts{}

	for auth := range registryAuthStore.RegistryAuths(ctx) {
		registryHosts[auth.Host] = auth.RegistryHost
	}

	for src, use := range o.HostAliases {
		for auth := range registryAuthStore.RegistryAuths(ctx) {
			if use == auth.Host {
				registryHosts[src] = auth.RegistryHost
			}
		}
	}

	if o.CacheDir != "" {
		return contentproxy.NewProxyFallbackRegistry(ctx,
			contentfs.NewNamespace(local.NewFS(o.CacheDir)),
			registryHosts,
		)
	}

	return contentremote.New(ctx, registryHosts)
}
