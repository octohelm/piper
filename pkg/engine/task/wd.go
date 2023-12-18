package task

import (
	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/piper/pkg/wd"
)

var WorkDirStore = &Store[string, wd.WorkDir]{}

var WorkDirContext = contextx.New[*Store[string, wd.WorkDir]](
	contextx.WithDefaultsFunc[*Store[string, wd.WorkDir]](func() *Store[string, wd.WorkDir] {
		return WorkDirStore
	}),
)
