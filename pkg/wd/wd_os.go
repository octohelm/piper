package wd

import (
	"context"
	"github.com/k0sproject/rig"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type CanOSInfo interface {
	OSInfo(ctx context.Context) (*OSInfo, error)
}

type OSInfo struct {
	rig.OSVersion

	Platform v1.Platform
}

var _ CanOSInfo = &wd{}

func (w *wd) OSInfo(ctx context.Context) (*OSInfo, error) {
	gnuarch, err := w.connection.ExecOutput("uname -m")
	if err != nil {
		return nil, err
	}
	os, _ := GoOS(w.connection.OSVersion.ID)
	arch, _ := GoArch(os, gnuarch)

	return &OSInfo{
		Platform: v1.Platform{
			OS:           os,
			Architecture: arch,
		},
		OSVersion: *w.connection.OSVersion,
	}, nil
}
