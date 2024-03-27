package task

import (
	contextx "github.com/octohelm/x/context"
	"sync"
)

var cached = &sync.Map{}

var CachedContext = contextx.New[Cached](contextx.WithDefaultsFunc(func() Cached {
	return cached
}))

type Cached interface {
	Load(k any) (any, bool)
	Store(k any, v any)
}
