package container

import (
	"context"

	"dagger.io/dagger"

	"github.com/octohelm/cuekit/pkg/cueflow/task"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Pull{})
}

// Pull
// image from
type Pull struct {
	task.Task

	// image from
	Source string `json:"source"`
	// image platform
	Platform string `json:"platform,omitzero" output:"platform"`
	// registry auth
	Auth *Auth `json:"auth,omitzero"`

	// image
	Output Container `json:"-" output:"output"`
}

func (x *Pull) Do(ctx context.Context) error {
	engine := piperdagger.Select(ctx, piperdagger.Scope{Platform: piperdagger.Platform(x.Platform)})

	requestPlatform := engine.Scope().Platform

	return engine.Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		dc := c.Container(dagger.ContainerOpts{Platform: requestPlatform})

		dc = RegistryAuthStoreContext.From(ctx).ApplyTo(ctx, c, dc, x.Source, x.Auth)

		dc = dc.From(x.Source)

		return x.Output.Sync(ctx, dc, string(requestPlatform))
	})
}
