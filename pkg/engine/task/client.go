package task

import (
	"context"

	contextx "github.com/octohelm/x/context"

	"github.com/octohelm/piper/pkg/wd"
)

var ClientContext = contextx.New[Client]()

type Client interface {
	SourceDir(ctx context.Context) (wd.WorkDir, error)
	CacheDir(ctx context.Context, tpe string) (wd.WorkDir, error)
}
