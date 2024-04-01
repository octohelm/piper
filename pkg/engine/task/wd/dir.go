package wd

import (
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

type Dir struct {
	task.Checkpoint

	// current work dir
	WorkDir wd.WorkDir `json:"wd"`
	// path related from current work dir
	Path string `json:"path"`
}

var _ cueflow.OutputValuer = &Dir{}

func (f *Dir) OutputValues() map[string]any {
	return map[string]any{
		"wd":   f.WorkDir,
		"path": f.Path,
	}
}
