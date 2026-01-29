package container

import (
	"context"
	"fmt"
	"iter"
	"net"
	"net/url"
	"strings"
	"sync"

	"dagger.io/dagger"
	"github.com/octohelm/crkit/pkg/content/remote"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/piper/pkg/engine/task/client"
)

func Secret(ctx context.Context, c *dagger.Client, s *client.Secret) (*dagger.Secret, bool) {
	if v, ok := s.Value(ctx); ok {
		return c.SetSecret(s.Ref.ID, v.Value), true
	}
	return nil, false
}

type Auth struct {
	Server   string        `json:"server,omitzero"`
	Username string        `json:"username"`
	Secret   client.Secret `json:"secret"`
}

func (a *Auth) ApplyTo(ctx context.Context, c *dagger.Client, container *dagger.Container, address string) *dagger.Container {
	if val, ok := a.Secret.Value(ctx); ok {
		secret := c.SetSecret(a.Secret.Ref.ID, val.Value)
		container = container.WithRegistryAuth(address, a.Username, secret)
	}
	return container
}

var defaultRegistryStore = NewRegistryAuthStore()

var RegistryAuthStoreContext = contextx.New[RegistryAuthStore](
	contextx.WithDefaultsFunc(func() RegistryAuthStore {
		return defaultRegistryStore
	}),
)

type RegistryAuthStore interface {
	Store(address string, auth *Auth)
	RegistryAuths(ctx context.Context) iter.Seq[RegistryAuth]
	ApplyTo(ctx context.Context, c *dagger.Client, container *dagger.Container, address string, localAuths ...*Auth) *dagger.Container
}

type RegistryAuth struct {
	Host string

	remote.RegistryHost
}

func NewRegistryAuthStore() RegistryAuthStore {
	return &registryAuthStore{}
}

type registryAuthStore struct {
	m sync.Map
}

func (r *registryAuthStore) RegistryAuths(ctx context.Context) iter.Seq[RegistryAuth] {
	return func(yield func(RegistryAuth) bool) {
		for key, value := range r.m.Range {
			host := key.(string)
			a := value.(*Auth)

			if password, ok := a.Secret.Value(ctx); ok {
				ra := RegistryAuth{
					Host: host,
				}

				ra.Server = a.Server

				if ra.Server == "" {
					u, err := url.Parse(fmt.Sprintf("https://%s", host))
					if err == nil {
						// ip hostname http for defaults
						if ip := net.ParseIP(u.Hostname()); ip != nil {
							u.Scheme = "http"
						}
						ra.Server = u.String()
					}
				}

				ra.Auth = &remote.RegistryAuth{
					Username: a.Username,
					Password: password.Value,
				}

				if !yield(ra) {
				}
			}
		}
	}
}

func (r *registryAuthStore) Store(address string, auth *Auth) {
	r.m.Store(address, auth)
}

func (r *registryAuthStore) ApplyTo(ctx context.Context, c *dagger.Client, container *dagger.Container, imageAddress string, localAuths ...*Auth) *dagger.Container {
	for key, value := range r.m.Range {
		domain := key.(string)

		if strings.HasPrefix(imageAddress, domain+"/") {
			a := value.(*Auth)
			container = a.ApplyTo(ctx, c, container, domain)
		}
	}

	for _, a := range localAuths {
		if a != nil {
			container = a.ApplyTo(ctx, c, container, imageAddress)
		}
	}

	return container
}
