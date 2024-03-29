package task

import (
	"github.com/octohelm/piper/pkg/generic/record"
	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/piper/pkg/wd"
)

var WorkDirStore = &record.Map[string, wd.WorkDir]{}

var WorkDirContext = contextx.New[*record.Map[string, wd.WorkDir]](
	contextx.WithDefaultsFunc[*record.Map[string, wd.WorkDir]](func() *record.Map[string, wd.WorkDir] {
		return WorkDirStore
	}),
)
