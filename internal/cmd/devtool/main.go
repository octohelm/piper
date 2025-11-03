package main

import (
	"context"
	"os"

	"github.com/innoai-tech/infra/devpkg/gengo"
	"github.com/innoai-tech/infra/pkg/cli"
	"github.com/octohelm/gengo/pkg/format"
)

import (
	_ "github.com/octohelm/enumeration/devpkg/enumgen"
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

var App = cli.NewApp("gengo", "dev")

func init() {
	cli.AddTo(App, &struct {
		cli.C `name:"gen"`
		gengo.Gengo
	}{})

	cli.AddTo(App, &struct {
		cli.C `name:"fmt"`
		format.Project
	}{})
}

func main() {
	if err := cli.Execute(context.Background(), App, os.Args[1:]); err != nil {
		panic(err)
	}
}
