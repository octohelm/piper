package wd

import (
	"context"
	"github.com/k0sproject/rig"
)

type CanOSInfo interface {
	OSInfo(ctx context.Context) (*OSInfo, error)
}

type OSInfo struct {
	rig.OSVersion

	Platform Platform

	Home string
}

var _ CanOSInfo = &wd{}

func (w *wd) OSInfo(ctx context.Context) (*OSInfo, error) {
	gnuarch, err := w.connection.ExecOutput("uname -m")
	if err != nil {
		return nil, err
	}
	os, _ := GoOS(w.connection.OSVersion.ID)
	arch, _ := GoArch(os, gnuarch)

	home, err := w.connection.ExecOutput("printenv HOME")
	if err != nil {
		return nil, err
	}

	return &OSInfo{
		Platform: Platform{
			OS:           os,
			Architecture: arch,
		},
		OSVersion: *w.connection.OSVersion,
		Home:      home,
	}, nil
}
