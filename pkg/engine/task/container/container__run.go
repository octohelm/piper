package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Run{})
}

type Run struct {
	task.Task

	Input Container `json:"input"`

	Mounts  map[string]Mount                 `json:"mounts,omitempty"`
	Env     map[string]client.SecretOrString `json:"env,omitempty"`
	Workdir string                           `json:"workdir,omitempty" default:"/"`
	User    string                           `json:"user,omitempty" default:"root:root"`
	Always  bool                             `json:"always,omitempty"`

	Shell string               `json:"shell,omitempty" default:"sh"`
	Run   client.StringOrSlice `json:"run"`

	Output Container `json:"-" output:"output"`
}

func (x *Run) Do(ctx context.Context) error {
	exec := &Exec{}

	exec.Input = x.Input
	exec.Mounts = x.Mounts
	exec.Env = x.Env
	exec.Workdir = x.Workdir
	exec.User = x.User
	exec.Always = x.Always

	if strings.HasPrefix(x.Shell, "/") {
		exec.Entrypoint = []string{x.Shell}
	} else {
		exec.Entrypoint = []string{fmt.Sprintf("/bin/%s", x.Shell)}
	}
	exec.Args = []string{
		"-c", strings.Join(x.Run, "\n"),
	}

	if err := exec.Do(ctx); err != nil {
		return err
	}

	x.Output = exec.Output

	return nil
}
