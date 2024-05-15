package task

import (
	"os"

	"github.com/k0sproject/rig"
)

func init() {
	if os.Getenv("ENABLE_RIG_LOG") != "1" {
		rig.SetLogger(&discord{})
	}
}

type discord struct{}

func (discord) Tracef(s string, i ...interface{}) {
}

func (discord) Debugf(s string, i ...interface{}) {
}

func (discord) Infof(s string, i ...interface{}) {
}

func (discord) Warnf(s string, i ...interface{}) {
}

func (discord) Errorf(s string, i ...interface{}) {
}
