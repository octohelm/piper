package client

import (
	"context"
	"github.com/octohelm/cuekit/pkg/version/gomod"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/wd"

	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &RevInfo{})
}

type RevInfo struct {
	task.SetupTask

	// get pseudo version same of go mod
	// like
	//   v0.0.0-20231222030512-c093d5e89975
	//   v0.0.0-dirty.0.20231222022414-5f9d1d44dacc
	Version string `json:"-" output:"version"`
}

func (t *RevInfo) ResultValue() any {
	return map[string]any{
		"version": t.Version,
	}
}

func (t *RevInfo) Do(ctx context.Context) error {
	cwd, err := task.ClientContext.From(ctx).SourceDir(ctx)
	if err != nil {
		return err
	}
	realPath, err := wd.RealPath(cwd)
	if err != nil {
		return err
	}
	r, err := gomod.RevInfoFromDir(context.Background(), realPath)
	if err != nil {
		return err
	}
	t.Version = r.Version
	return nil
}
