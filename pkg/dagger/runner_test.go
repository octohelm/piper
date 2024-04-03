package dagger

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"dagger.io/dagger"

	testingx "github.com/octohelm/x/testing"
)

func TestEngine(t *testing.T) {
	t.Run("simple flow", func(t *testing.T) {
		ctx := context.Background()

		r, err := NewRunner()
		testingx.Expect(t, err, testingx.Be[error](nil))
		defer r.Shutdown(ctx)

		engine := r.Select(ctx, Scope{
			Platform: Platform(fmt.Sprintf("linux/%s", runtime.GOARCH)),
		})
		defer engine.Shutdown(ctx)

		var dirID dagger.DirectoryID

		err = engine.Do(ctx, func(ctx context.Context, c *Client) error {
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
		testingx.Expect(t, err, testingx.Be[error](nil))

		err = engine.Do(ctx, func(ctx context.Context, c *Client) error {
			loaded := c.Container().WithRootfs(c.LoadDirectoryFromID(dirID))
			_, err = loaded.Directory("/dist").Export(ctx, "output")
			return err
		})
		testingx.Expect(t, err, testingx.Be[error](nil))
	})
}
