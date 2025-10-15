package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"cuelang.org/go/cue"
	"github.com/innoai-tech/infra/pkg/cli"
)

func init() {
	cli.AddTo(App, &Version{})
}

type Version struct {
	cli.C

	VersionRun
}

type VersionRun struct {
}

func (r *VersionRun) Run(ctx context.Context) error {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("unknown error reading build-info")
	}

	w := os.Stdout
	_, _ = fmt.Fprintf(w, "cue version %s\n\n", cue.LanguageVersion())
	_, _ = fmt.Fprintf(w, "CUE language version %s\n\n", cue.LanguageVersion())
	_, _ = fmt.Fprintf(w, "Go version %s\n", runtime.Version())

	for _, s := range bi.Settings {
		if s.Value == "" {
			continue
		}
		_, _ = fmt.Fprintf(w, "%16s %s\n", s.Key, s.Value)
	}

	return nil
}
