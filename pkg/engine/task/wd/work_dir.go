package wd

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

type WorkDir struct {
	Ref struct {
		ID string `json:"id"`
	} `json:"$$wd"`
}

func (w *WorkDir) Get(ctx context.Context, optFns ...wd.OptionFunc) (wd.WorkDir, error) {
	if found, ok := task.WorkDirContext.From(ctx).Get(w.Ref.ID); ok {
		return wd.With(found, optFns...)
	}
	return nil, errors.Errorf("workdir %s is not found", w.Ref.ID)
}

func (w *WorkDir) Do(ctx context.Context, action func(ctx context.Context, wd wd.WorkDir) error, optFns ...wd.OptionFunc) error {
	cwd, err := w.Get(ctx, optFns...)
	if err != nil {
		return err
	}
	if err := action(ctx, cwd); err != nil {
		return errors.Wrapf(err, "%s", cwd)
	}
	return nil
}

func (w *WorkDir) SetBy(ctx context.Context, workdir wd.WorkDir) {
	w.Ref.ID = task.WorkDirContext.From(ctx).Set(workdir)
}

func (w *WorkDir) ScopeName(ctx context.Context) (string, error) {
	cwd, err := w.Get(ctx)
	if err != nil {
		return "", err
	}
	return cwd.String(), nil
}

type CurrentWorkDir struct {
	// current word dir
	Cwd WorkDir `json:"cwd"`
}

var _ cueflow.WithScopeName = CurrentWorkDir{}

func (w CurrentWorkDir) ScopeName(ctx context.Context) (string, error) {
	return w.Cwd.ScopeName(ctx)
}
