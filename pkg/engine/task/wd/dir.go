package wd

import (
	"github.com/octohelm/piper/pkg/wd"
)

type Dir struct {
	// current work dir
	WorkDir wd.WorkDir `json:"wd"`
	// path related from current work dir
	Path string `json:"path"`
}
