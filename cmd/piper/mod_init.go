package main

import (
	"context"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/cuekit/pkg/cuecontext"
	"github.com/pkg/errors"
	"os"
)

func init() {
	cli.AddTo(Mod, &Init{})
}

type Init struct {
	cli.C
	InitRun
}

type InitRun struct {
	Name []string `arg:""`
}

func (r *InitRun) Run(ctx context.Context) error {
	if len(r.Name) > 0 && r.Name[0] != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return cuecontext.Init(ctx, cwd, r.Name[0])
	}

	return errors.New("name is required")
}
