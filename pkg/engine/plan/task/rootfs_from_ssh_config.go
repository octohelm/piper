package task

import (
	"context"
	"github.com/k0sproject/rig"
	"github.com/octohelm/piper/pkg/engine/plan/task/core"
	"github.com/octohelm/piper/pkg/engine/rigutil"
	"github.com/octohelm/piper/pkg/sshutil"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	core.DefaultFactory.Register(&RootfsFromSSHConfig{})
}

type RootfsFromSSHConfig struct {
	core.SetupTask

	Config string `json:"config" default:"~/.ssh/config"`
	Alias  string `json:"alias"`

	WorkDir core.WD `json:"-" piper:"generated,name=wd"`
}

func (c *RootfsFromSSHConfig) Do(ctx context.Context) error {
	ssh, err := sshutil.Load(c.Config, c.Alias)
	if err != nil {
		return err
	}

	cwd, err := wd.Wrap(
		&rig.Connection{
			SSH: ssh,
		},
		wd.WithUser(ssh.User),
	)
	if err != nil {
		return err
	}

	c.WorkDir = *core.WDOfID(rigutil.WorkDirContext.From(ctx).Set(cwd))
	return nil
}
