package main

import (
	"context"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/cuekit/pkg/cuecontext"
	"os"
)

func init() {
	cli.AddTo(Mod, &Tidy{})
}

type Tidy struct {
	cli.C
	TidyRun
}

type TidyRun struct {
}

func (r *TidyRun) Run(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return cuecontext.Tidy(ctx, cwd)
}
