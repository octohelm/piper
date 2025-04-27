package wd

import (
	"context"

	"github.com/k0sproject/rig"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &SSH{})
}

// SSH
// create ssh work dir for remote executing
type SSH struct {
	task.Task
	// ssh address
	Address string `json:"address"`
	// ssh hostKey
	HostKey string `json:"hostKey,omitzero"`
	// ssh identity file
	IdentityFile string `json:"identityFile"`
	// ssh port
	Port int `json:"port,omitzero" default:"22"`
	// ssh user
	User string `json:"user,omitzero" default:"root"`
	// the workdir from ssh
	WorkDir WorkDir `json:"-" output:"dir"`
}

func (c *SSH) Do(ctx context.Context) error {
	conn := &rig.Connection{}
	conn.SSH = &rig.SSH{
		Address: c.Address,
		Port:    c.Port,
		User:    c.User,
		HostKey: c.HostKey,
		KeyPath: &c.IdentityFile,
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

	c.WorkDir.Ref.ID = cwd.Addr().String()

	task.WorkDirContext.From(ctx).Store(c.WorkDir.Ref.ID, cwd)

	return nil
}
