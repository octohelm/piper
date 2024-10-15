package wd

import (
	"context"
	"fmt"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

type WorkDir struct {
	Ref struct {
		ID string `json:"id"`
	} `json:"$$wd"`
}

func (w *WorkDir) Get(ctx context.Context, optFns ...wd.OptionFunc) (wd.WorkDir, error) {
	if found, ok := task.WorkDirContext.From(ctx).Load(w.Ref.ID); ok {
		return wd.With(found, optFns...)
	}
	return nil, fmt.Errorf("workdir %s is not found", w.Ref.ID)
}

func (w *WorkDir) Do(ctx context.Context, action func(ctx context.Context, wd wd.WorkDir) error, optFns ...wd.OptionFunc) error {
	cwd, err := w.Get(ctx, optFns...)
	if err != nil {
		return err
	}
	if err := action(ctx, cwd); err != nil {
		return fmt.Errorf("do failed at %s: %w", cwd, err)
	}
	return nil
}

func (w *WorkDir) Sync(ctx context.Context, workdir wd.WorkDir) error {
	w.Ref.ID = workdir.Addr().String()
	task.WorkDirContext.From(ctx).Store(w.Ref.ID, workdir)
	return nil
}

func (w *WorkDir) ScopeName(ctx context.Context) (string, error) {
	cwd, err := w.Get(ctx)
	if err != nil {
		return "", err
	}
	return cwd.String(), nil
}
