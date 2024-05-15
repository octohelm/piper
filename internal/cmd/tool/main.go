package main

import (
	"context"
	"os"

	"github.com/innoai-tech/infra/pkg/otel"

	"github.com/innoai-tech/infra/devpkg/gengo"
	"github.com/innoai-tech/infra/pkg/cli"

	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/devpkg/enumgen"
)

var App = cli.NewApp("gengo", "dev")

func init() {
	c := &struct {
		cli.C `name:"gen"`
		otel.Otel
		gengo.Gengo
	}{}

	c.LogLevel = otel.DebugLevel

	cli.AddTo(App, c)
}

func main() {
	if err := cli.Execute(context.Background(), App, os.Args[1:]); err != nil {
		panic(err)
	}
}
