package main

import (
	"github.com/innoai-tech/infra/pkg/cli"

	"github.com/octohelm/piper/pkg/engine"
	"github.com/octohelm/piper/pkg/logutil"
)

func init() {
	cli.AddTo(App, &Do{})
}

type Do struct {
	cli.C
	logutil.Logger
	engine.Pipeline
}
