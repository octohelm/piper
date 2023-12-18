package main

import (
	"context"
	"os"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/innoai-tech/infra/devpkg/gengo"
	"github.com/innoai-tech/infra/pkg/cli"

	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/devpkg/enumgen"
)

var App = cli.NewApp("gengo", "dev")

func init() {
	cli.AddTo(App, &struct {
		cli.C `name:"gen"`
		gengo.Gengo
	}{})
}

func main() {
	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := cli.Execute(ctx, App, os.Args[1:]); err != nil {
		panic(err)
	}
}
