package core

import (
	"context"
	"github.com/octohelm/piper/pkg/engine/rigutil"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
)

func init() {
	DefaultFactory.Register(&WD{})
}

type WD struct {
	Meta struct {
		Wd struct {
			ID string `json:"id,omitempty"`
		} `json:"wd"`
	} `json:"$piper"`
}

func (w *WD) getWd(ctx context.Context, optFns ...wd.OptionFunc) (wd.WorkDir, error) {
	if found, ok := rigutil.WorkDirContext.From(ctx).Get(w.Meta.Wd.ID); ok {
		return wd.With(found, optFns...)
	}
	return nil, errors.Errorf("workdir %s is not found", w.Meta.Wd.ID)
}

func (w *WD) Do(ctx context.Context, action func(wd wd.WorkDir) error, optFns ...wd.OptionFunc) error {
	rootfs, err := w.getWd(ctx, optFns...)
	if err != nil {
		return err
	}
	return action(rootfs)
}

func (w *WD) SetBy(ctx context.Context, workdir wd.WorkDir) {
	w.Meta.Wd.ID = rigutil.WorkDirContext.From(ctx).Set(workdir)
}

func WDOfID(id string) *WD {
	fs := &WD{}
	fs.Meta.Wd.ID = id
	return fs
}
