package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"dagger.io/dagger"
	pkgdagger "github.com/octohelm/piper/pkg/dagger"
)

const InstrumentationLibrary = "piper/cli"

func main() {
	_ = os.Setenv("PIPER_BUILDER_HOST", "")

	err := pkgdagger.Run(context.Background(), "debug", func(ctx context.Context) (rerr error) {
		engine := pkgdagger.RunnerContext.From(ctx).Select(ctx, pkgdagger.Scope{
			Platform: pkgdagger.Platform(fmt.Sprintf("linux/%s", runtime.GOARCH)),
		})

		var dirID dagger.DirectoryID

		err := engine.Do(ctx, func(ctx context.Context, c *pkgdagger.Client) error {
			container := c.Container().
				From("busybox").
				WithEnvVariable("DATE", time.Now().String()).
				WithExec([]string{"sh", "-c", "mkdir -p /dist"}).
				WithExec([]string{"sh", "-c", "echo ${DATE} > /dist/txt"})

			d, err := container.Rootfs().Sync(ctx)
			if err != nil {
				return err
			}
			id, err := d.ID(ctx)
			if err != nil {
				return err
			}
			dirID = id
			return nil
		})

		if err != nil {
			return err
		}

		return engine.Do(ctx, func(ctx context.Context, c *pkgdagger.Client) error {
			loaded := c.Container().WithRootfs(c.LoadDirectoryFromID(dirID))
			_, err = loaded.Directory("/dist").Export(ctx, "target")
			return err
		})
	})

	if err != nil {
		panic(err)
	}
}
