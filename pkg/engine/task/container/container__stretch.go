package container

import (
	"context"
	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Stretch{})
}

// Stretch
// image from stretch
type Stretch struct {
	task.Task
	// image platform
	Platform string `json:"platform,omitempty" output:"platform"`
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
