package task

import (
	"context"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	core.DefaultFactory.Register(&SysInfo{})
}

type SysInfo struct {
	core.Task

	CWD core.WD `json:"cwd"`

	Release  Release  `json:"-" piper:"generated,name=release"`
	Platform Platform `json:"-" piper:"generated,name=platform"`
}

type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

type Release struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	IDLike  string `json:"id_like,omitempty"`
	Version string `json:"version,omitempty"`
}

func (e *SysInfo) Do(ctx context.Context) error {
	return e.CWD.Do(ctx, func(cwd wd.WorkDir) error {
		if can, ok := cwd.(wd.CanOSInfo); ok {
			info, err := can.OSInfo(ctx)
			if err != nil {
				return err
			}

			e.Release.Name = info.Name
			e.Release.ID = info.ID
			e.Release.IDLike = info.IDLike
			e.Release.Version = info.Version

			e.Platform.OS = info.Platform.OS
			e.Platform.Architecture = info.Platform.Architecture
		}
		return nil
	})
}
