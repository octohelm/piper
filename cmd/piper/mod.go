package main

import (
	"github.com/innoai-tech/infra/pkg/cli"
)

var Mod = cli.AddTo(App, &struct {
	cli.C `name:"mod"`
}{})
