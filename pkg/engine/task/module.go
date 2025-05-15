package task

import (
	"github.com/octohelm/cuekit/pkg/mod/module"
	contextx "github.com/octohelm/x/context"
)

var ModuleContext = contextx.New[*module.Module]()
