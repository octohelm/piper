package processpool

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func Test_processPool(t *testing.T) {
	p := NewProcessPool("test")

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	go func() {
		p.Wait(ctx)
	}()
	defer p.Close()

	repo, _ := name.NewRepository("docker.io/a/b")

	wg := &sync.WaitGroup{}

	for i := range 5 {
		u := p.Progress(repo.Tag(fmt.Sprintf("v%d", i)))

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(u)

			for j := range 10 {
				u <- v1.Update{
					Complete: int64(j),
					Total:    10,
				}
			}
		}()
	}

	wg.Wait()
}
