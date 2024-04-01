package container

import (
	"context"
	"github.com/go-courier/logr"
	"sync"

	"dagger.io/dagger"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"

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

func (container *Container) ContainerID() dagger.ContainerID {
	if k, ok := containerIDs.Load(container.Ref.ID); ok {
		return k.(containerMeta).id
	}
	return ""
}

func (container *Container) Container(c *dagger.Client) *dagger.Container {
	if id := container.ContainerID(); id != "" {
		return c.LoadContainerFromID(id)
	}
	return c.Container()
}

func (container *Container) Sync(ctx context.Context, c *dagger.Container, platform string) error {
	cc, err := c.Sync(ctx)
	if err != nil {
		return err
	}
	id, err := cc.ID(ctx)
	if err != nil || id == "" {
		return errors.Wrap(err, "resolve container id failed")
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
