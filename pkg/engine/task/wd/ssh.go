package wd

import (
	"context"

	"github.com/k0sproject/rig"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/sshutil"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &SSH{})
}

// SSH
// create ssh work dir for remote executing
type SSH struct {
	task.SetupTask

	// path to ssh config
	Config string `json:"config,omitempty"`
	// host key of ssh config
	HostKey string `json:"hostKey,omitempty"`

	// custom setting
	// ssh address
	Address string `json:"address,omitempty"`
	// ssh port
	Port int `json:"port,omitempty" default:"22"`
	// ssh identity file
	IdentityFile string `json:"identityFile,omitempty"`
	// ssh user
	User string `json:"user,omitempty" default:"root"`

	WorkDir WorkDir `json:"-" output:"wd"`
}

func (c *SSH) ResultValue() any {
	return c.WorkDir
}

func (c *SSH) Do(ctx context.Context) error {
	conn := &rig.Connection{}

	if c.Config != "" {
		ssh, err := sshutil.Load(c.Config, c.HostKey)
		if err != nil {
			return err
		}
		conn.SSH = ssh
	} else {
		conn.SSH = &rig.SSH{
			Address: c.Address,
			Port:    c.Port,
			User:    c.User,
			KeyPath: &c.IdentityFile,
		}
	}

	user := conn.SSH.User

	if c.User != "" {
		user = c.User
	}

	cwd, err := wd.Wrap(
		conn,
		wd.WithUser(user),
	)

	if err != nil {
		return err
	}

	c.WorkDir.Ref.ID = task.WorkDirContext.From(ctx).Set(cwd)

	return nil
}
