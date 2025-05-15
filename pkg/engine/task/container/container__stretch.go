package container

import (
	"context"

	"dagger.io/dagger"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&Stretch{})
}

// Stretch
// image from stretch
type Stretch struct {
	task.Task
	// image platform
	Platform string `json:"platform,omitzero" output:"platform"`
	// image
	Output Container `json:"-" output:"output"`
}

func (x *Stretch) Do(ctx context.Context) error {
	engine := piperdagger.Select(ctx, piperdagger.Scope{Platform: piperdagger.Platform(x.Platform)})
	requestPlatform := engine.Scope().Platform

	return engine.Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		dc := c.Container(dagger.ContainerOpts{Platform: requestPlatform})
		return x.Output.Sync(ctx, dc, string(requestPlatform))
	})
}
