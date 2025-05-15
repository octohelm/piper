package wd

import (
	"context"

	"github.com/k0sproject/rig"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/sshutil"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&SSHFromConfig{})
}

// SSHFromConfig
// create ssh work dir for remote executing
type SSHFromConfig struct {
	task.Task

	// path to ssh config
	Config string `json:"config"`
	// host key of ssh config
	HostKey string `json:"hostKey"`

	// the workdir from ssh
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (c *SSHFromConfig) Do(ctx context.Context) error {
	conn := &rig.Connection{}

	ssh, err := sshutil.Load(c.Config, c.HostKey)
	if err != nil {
		return err
	}
	conn.SSH = ssh

	user := conn.SSH.User

	cwd, err := wd.Wrap(
		conn,
		wd.WithUser(user),
	)
	if err != nil {
		return err
	}

	c.WorkDir.Ref.ID = cwd.Addr().String()

	enginetask.WorkDirContext.From(ctx).Store(c.WorkDir.Ref.ID, cwd)

	return nil
}
