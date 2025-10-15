package main

import (
	"context"

	"cuelang.org/go/cue/lsp"
	"cuelang.org/go/mod/modconfig"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/cuekit/pkg/cuecontext"
)

func init() {
	cli.AddTo(App, &Lsp{})
}

type Lsp struct {
	cli.C

	LspRun
}

type LspRun struct {
	Args []string `arg:""`
}

func (r *LspRun) Run(ctx context.Context) error {
	lsp.SetDefaultRegistryFunc(func() (modconfig.Registry, error) {
		c, err := cuecontext.NewConfig()
		if err != nil {
			return nil, err
		}
		return c.Registry, nil
	})

	lsp.Run(ctx, r.Args)
	return nil
}
