package task

import (
	"context"

	"github.com/octohelm/piper/pkg/wd"
	contextx "github.com/octohelm/x/context"
)

var ClientContext = contextx.New[Client]()

type Client interface {
	SourceDir(ctx context.Context) (wd.WorkDir, error)
	CacheDir(ctx context.Context, tpe string) (wd.WorkDir, error)
}
