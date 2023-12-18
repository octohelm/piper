package main

import (
	"context"
	"os"

	"github.com/innoai-tech/infra/pkg/cli"

	"github.com/octohelm/piper/internal/version"
)

var App = cli.NewApp("piper", version.Version())

func main() {
	if err := cli.Execute(context.Background(), App, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
