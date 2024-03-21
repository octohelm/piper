package container

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"sync"
)

var fsIDs = sync.Map{}

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
	if k, ok := fsIDs.Load(fs.Ref.ID); ok {
		if dc, ok := k.(fsContext); ok {
			return dc.id
		}
	}
	return ""
}

func (fs *Fs) Directory(ctx context.Context, c *dagger.Client) (*dagger.Directory, error) {
	if k, ok := fsIDs.Load(fs.Ref.ID); ok {
		switch x := k.(type) {
		case fsContext:
			return c.LoadDirectoryFromID(x.id), nil
		case lazyAction:
			return x.do(
				cueflow.TaskPathContext.Inject(ctx, x.taskPath),
				c.Pipeline(x.taskPath),
			)
		}
	}
	return c.Directory(), nil
}

func (fs *Fs) Sync(ctx context.Context, c *dagger.Directory) error {
	cc, err := c.Sync(ctx)
	if err != nil {
		return err
	}
	id, err := cc.ID(ctx)
	if err != nil || id == "" {
		return errors.Wrap(err, "resolve fs id failed")
	}

	key := "fs://" + digest.FromString(string(id)).String()
	fsIDs.Store(key, fsContext{
		id:    id,
		scope: piperdagger.ScopeContext.From(ctx),
	})
	fs.Ref.ID = key
	return nil
}

func (fs *Fs) SyncLazyDirectory(ctx context.Context, do LazyDoFn) error {
	key := "lazy://" + digest.FromString(fmt.Sprintf("%p", do)).String()
	fsIDs.Store(key, lazyAction{
		taskPath: cueflow.TaskPathContext.From(ctx),
		do:       do,
	})
	fs.Ref.ID = key
	return nil
}

type fsContext struct {
	id    dagger.DirectoryID
	scope piperdagger.Scope
}

type LazyDoFn func(ctx context.Context, c *dagger.Client) (*dagger.Directory, error)

type lazyAction struct {
	taskPath string
	do       LazyDoFn
}
