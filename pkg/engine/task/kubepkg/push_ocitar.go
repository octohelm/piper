package kubepkg

import (
	"context"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	taskocitar "github.com/octohelm/piper/pkg/engine/task/ocitar"
)

func init() {
	cueflow.RegisterTask(task.Factory, &PushOciTar{})
}

type PushOciTar struct {
	task.Task

	// SrcFile of oci tar
	SrcFile file.File `json:"srcFile"`

	// RemoteURL of container registry
	RemoteURL string `json:"remoteURL"`

	// Rename for image repo name
	// go template rule
	// `{{ .registry }}/{{ .namespace }}/{{ .name }}`
	Rename taskocitar.Rename `json:"rename,omitempty"`
}

func (t *PushOciTar) Do(ctx context.Context) error {
	p := taskocitar.Push{}
	p.Task = t.Task

	p.RemoteURL = t.RemoteURL
	p.SrcFile = t.SrcFile
	p.Rename = t.Rename

	return p.Do(ctx)
}
