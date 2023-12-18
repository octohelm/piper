package wd

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &SysInfo{})
}

// SysInfo
// get sys info of current work dir
type SysInfo struct {
	task.Task

	CurrentWorkDir

	// home
	Home string `json:"-" output:"home"`
	// os release info
	Release Release `json:"-" output:"release"`
	// os platform
	Platform Platform `json:"-" output:"platform"`
}

func (t SysInfo) ResultValue() any {
	return map[string]any{
		"home":     t.Home,
		"release":  t.Release,
		"platform": t.Platform,
	}
}

type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

type Release struct {
	// OS name
	Name string `json:"name"`
	// OS Version
	Version string `json:"version,omitempty"`
	// OS id, like `ubuntu` `windows`
	ID string `json:"id"`
	// if os is based on some upstream
	// like debian when id is `ubuntu`
	IDLike string `json:"id_like,omitempty"`
}

func (t *SysInfo) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) error {
		if can, ok := cwd.(wd.CanOSInfo); ok {
			info, err := can.OSInfo(ctx)
			if err != nil {
				return err
			}

			t.Release.Name = info.Name
			t.Release.ID = info.ID
			t.Release.IDLike = info.IDLike
			t.Release.Version = info.Version

			t.Platform.OS = info.Platform.OS
			t.Platform.Architecture = info.Platform.Architecture

			t.Home = info.Home
		}
		return nil
	})
}
