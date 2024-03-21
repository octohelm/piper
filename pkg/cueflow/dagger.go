package cueflow

import (
	"context"
	"github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/x/ptr"
	"github.com/vito/progrock"
	"strings"
)

func WrapDaggerRunner(r dagger.Runner) dagger.Runner {
	return &runnerWrapper{
		underlying: r,
	}
}

type runnerWrapper struct {
	underlying dagger.Runner
}

func (r *runnerWrapper) Select(ctx context.Context, scope dagger.Scope) dagger.Engine {
	return &engineWrapper{Engine: r.underlying.Select(ctx, scope)}
}

func (r *runnerWrapper) Shutdown(ctx context.Context) error {
	return r.underlying.Shutdown(ctx)
}

type engineWrapper struct {
	dagger.Engine
}

func (e *engineWrapper) Do(ctx context.Context, do func(ctx context.Context, client *dagger.Client) error) error {
	return e.Engine.Do(ctx, func(ctx context.Context, client *dagger.Client) error {
		return do(ctx, client.Pipeline(encodeTaskPath(TaskPathContext.From(ctx))))
	})
}

func WrapProgrockWriter(w progrock.Writer) progrock.Writer {
	return &customProgrockWriter{
		Writer: w,
	}
}

type customProgrockWriter struct {
	progrock.Writer
}

func isTaskPath(id string) bool {
	return strings.HasPrefix(id, "~")
}

func encodeTaskPath(s string) string {
	return "~" + s
}

func decodeTaskPath(id string) string {
	return strings.Split(id[1:], "@")[0]
}

func (c *customProgrockWriter) WriteStatus(update *progrock.StatusUpdate) error {
	for _, g := range update.Groups {
		if isTaskPath(g.Id) {
			g.Id = decodeTaskPath(g.Id)
		}

		if isTaskPath(g.Name) {
			g.Name = decodeTaskPath(g.Name)
		}

		if g.Parent != nil {
			parent := *g.Parent

			if isTaskPath(parent) {
				g.Parent = ptr.String(decodeTaskPath(parent))
			}
		}
	}

	for _, m := range update.Memberships {
		if isTaskPath(m.Group) {
			m.Group = decodeTaskPath(m.Group)
		}
	}

	return c.Writer.WriteStatus(update)
}
