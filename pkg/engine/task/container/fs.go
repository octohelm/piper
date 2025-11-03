package container

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger.io/dagger"
	"github.com/opencontainers/go-digest"

	"github.com/octohelm/cuekit/pkg/cueflow"
	"github.com/octohelm/x/logr"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/generic/record"
)

var fsIDs = record.Map[string, any]{}

type Fs struct {
	Ref struct {
		ID string `json:"id"`
	} `json:"$$fs"`
}

func (fs *Fs) Select(ctx context.Context) piperdagger.Engine {
	scope := fs.Scope()
	logr.FromContext(ctx).WithValues("fs", fs.Ref.ID, "scope", scope).Debug("selected engine")
	return piperdagger.RunnerContext.From(ctx).Select(ctx, scope)
}

func (fs *Fs) Scope() piperdagger.Scope {
	if k, ok := fsIDs.Load(fs.Ref.ID); ok {
		if dc, ok := k.(fsContext); ok {
			return dc.scope
		}
	}
	return piperdagger.Scope{}
}

func (fs *Fs) DirectoryID() dagger.DirectoryID {
	if value, ok := fsIDs.Load(fs.Ref.ID); ok {
		if dc, ok := value.(fsContext); ok {
			return dc.id
		}
	}
	return ""
}

func (fs *Fs) Directory(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
	if value, ok := fsIDs.Load(fs.Ref.ID); ok {
		switch x := value.(type) {
		case fsContext:
			return c.LoadDirectoryFromID(x.id), nil
		case lazyAction:
			nctx, l := logr.FromContext(ctx).Start(ctx, x.taskPath)
			defer l.End()

			return x.do(nctx, c)
		}
	}
	return nil, fmt.Errorf("fs is not found: %s", fs.Ref.ID)
}

func (fs *Fs) Sync(ctx context.Context, c *dagger.Directory) error {
	cc, err := c.Sync(ctx)
	if err != nil {
		return err
	}
	id, err := cc.ID(ctx)
	if err != nil || id == "" {
		return fmt.Errorf("resolve fs id failed: %w", err)
	}

	scope := piperdagger.ScopeContext.From(ctx)
	key := fmt.Sprintf("dagger-dir://%s?scope=%s", digest.FromString(string(id)), scope)
	fsIDs.Store(key, fsContext{
		id:    id,
		scope: scope,
	})
	fs.Ref.ID = key
	return nil
}

func (fs *Fs) SyncLazyDirectory(ctx context.Context, inputs any, do LazyDoFn) error {
	data, err := json.Marshal(inputs)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("lazy://%s", digest.FromBytes(data))

	_, loaded := fsIDs.LoadOrStore(key, lazyAction{
		taskPath: cueflow.TaskPathContext.From(ctx),
		do:       do,
	})

	if loaded {
		panic(fmt.Errorf("lazy fs conflicted, %s", key))
	}

	fs.Ref.ID = key

	return nil
}

type fsContext struct {
	scope piperdagger.Scope
	id    dagger.DirectoryID
}

type LazyDoFn func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error)

type lazyAction struct {
	taskPath string
	do       LazyDoFn
}
