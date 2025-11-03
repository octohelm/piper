package container

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/x/logr"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/generic/record"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Push{})
}

// Push image to registry
type Push struct {
	task.Task
	// image tag
	Dest string `json:"dest"`
	// images for push
	// [Platform]: _
	Images map[string]Container `json:"images"`
	// annotations
	Annotations map[string]string `json:"annotations,omitzero"`

	// registry auth
	Auth *Auth `json:"auth,omitzero"`

	// image pushed result
	Result string `json:"-" output:"result"`
}

func (x *Push) Do(ctx context.Context) error {
	if len(x.Images) == 0 {
		return errors.New("Push at least one image")
	}

	eg := &errgroup.Group{}
	published := record.Map[*pkgwd.Platform, string]{}

	for platform, container := range x.Images {
		eg.Go(func() error {
			return container.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
				cc, err := container.Container(ctx, c)
				if err != nil {
					return err
				}

				p, err := pkgwd.ParsePlatform(platform)
				if err != nil {
					return fmt.Errorf("parse platform failed: %s: %w", p, err)
				}

				// push without tag tmp
				image := fmt.Sprintf("%s:tmp", strings.Split(x.Dest, ":")[0])

				logr.FromContext(ctx).WithValues("platform", p).Info(fmt.Sprintf("publishing %s", image))

				cc = RegistryAuthStoreContext.From(ctx).ApplyTo(ctx, c, cc, image, x.Auth)

				tag, err := cc.Publish(ctx, image)
				if err != nil {
					return fmt.Errorf("published image failed")
				}

				published.Store(p, tag)
				return nil
			})
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return piperdagger.Select(ctx, piperdagger.Scope{}).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		opt := dagger.ContainerPublishOpts{}

		for platform, image := range published.Range {
			cc := c.Container(dagger.ContainerOpts{Platform: dagger.Platform(platform.String())})
			cc = RegistryAuthStoreContext.From(ctx).ApplyTo(ctx, c, cc, image, x.Auth)
			opt.PlatformVariants = append(opt.PlatformVariants, cc.From(image))
		}

		finalC := c.Container()

		if len(x.Annotations) > 0 {
			for k, v := range x.Annotations {
				finalC = finalC.WithAnnotation(k, v)
			}
		}

		result, err := finalC.Publish(ctx, x.Dest, opt)
		if err != nil {
			return fmt.Errorf("published manifest list failed")
		}

		logr.FromContext(ctx).WithValues("image", result).Info("published.")

		x.Result = result

		return nil
	})
}
