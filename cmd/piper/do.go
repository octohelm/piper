package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/piper/pkg/engine"
)

func init() {
	cli.AddTo(App, &Do{})
}

type Do struct {
	cli.C
	engine.Logger
	engine.Pipeline
}
