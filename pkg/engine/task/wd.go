package task

import (
	contextx "github.com/octohelm/x/context"
	syncx "github.com/octohelm/x/sync"

	"github.com/octohelm/piper/pkg/wd"
)

var WorkDirStore = &syncx.Map[string, wd.WorkDir]{}

var WorkDirContext = contextx.New[*syncx.Map[string, wd.WorkDir]](
	contextx.WithDefaultsFunc[*syncx.Map[string, wd.WorkDir]](func() *syncx.Map[string, wd.WorkDir] {
		return WorkDirStore
	}),
)
