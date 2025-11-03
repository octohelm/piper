package container

import (
	"context"
	"fmt"
	"sync"

	"dagger.io/dagger"
	"github.com/opencontainers/go-digest"

	"github.com/octohelm/x/logr"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
)

var containerIDs = sync.Map{}

type containerMeta struct {
	id    dagger.ContainerID
	scope piperdagger.Scope
}

type Container struct {
	Ref struct {
		ID string `json:"id"`
	} `json:"$$container"`
	Rootfs   Fs     `json:"rootfs"`
	Platform string `json:"platform"`
}

func (container *Container) Select(ctx context.Context) piperdagger.Engine {
	scope := container.Scope()
	logr.FromContext(ctx).WithValues("container", container.Ref.ID, "scope", scope).Debug("selected engine")
	return piperdagger.RunnerContext.From(ctx).Select(ctx, scope)
}

func (container *Container) Scope() piperdagger.Scope {
	if k, ok := containerIDs.Load(container.Ref.ID); ok {
		return k.(containerMeta).scope
	}
	return piperdagger.Scope{}
}

func (container *Container) Container(ctx context.Context, c *dagger.Client) (*dagger.Container, error) {
	if k, ok := containerIDs.Load(container.Ref.ID); ok {
		return c.LoadContainerFromID(k.(containerMeta).id), nil
	}
	return nil, fmt.Errorf("missing container %s", container.Ref.ID)
}

func (container *Container) Sync(ctx context.Context, c *dagger.Container, platform string) error {
	cc, err := c.Sync(ctx)
	if err != nil {
		return err
	}
	id, err := cc.ID(ctx)
	if err != nil || id == "" {
		return fmt.Errorf("resolve container id failed: %w", err)
	}

	if err := container.Rootfs.Sync(ctx, c.Rootfs()); err != nil {
		return err
	}

	container.Platform = platform

	container.storeContainerID(piperdagger.ScopeContext.From(ctx), id)

	return nil
}

func (container *Container) storeContainerID(scope piperdagger.Scope, id dagger.ContainerID) {
	key := digest.FromString(string(id)).String()

	containerIDs.Store(key, containerMeta{
		id:    id,
		scope: scope,
	})

	container.Ref.ID = key
}
