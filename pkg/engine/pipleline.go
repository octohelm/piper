package engine

import (
	"fmt"
	"os"

	cueerrors "cuelang.org/go/cue/errors"
	cuetoken "cuelang.org/go/cue/token"
	"golang.org/x/net/context"
)

type Pipeline struct {
	Action []string `arg:""`
	// plan root file
	// and the dir of the root file will be the cwd for all cue files
	Project string `flag:",omitzero" alias:"p"`

	// cache dir root
	// for cache files
	CacheDir string `flag:",omitzero"`
}

func (pipeline *Pipeline) SetDefaults() {
	if pipeline.Project == "" {
		pipeline.Project = "./piper.cue"
	}

	if pipeline.CacheDir == "" {
		pipeline.CacheDir = "~/.piper/cache"
	}
}

func (pipeline *Pipeline) Run(ctx context.Context) error {
	p, err := New(
		ctx,
		WithProject(pipeline.Project),
		WithCacheDir(pipeline.CacheDir),
	)
	if err != nil {
		return err
	}

	if err := p.Run(ctx, pipeline.Action...); err != nil {
		if errList := cueerrors.Errors(err); len(errList) > 0 {
			buf := os.Stderr

			records := map[cuetoken.Pos]bool{}

			for _, e := range errList {
				if _, ok := records[e.Position()]; !ok {
					cueerrors.Print(buf, e, nil)
					records[e.Position()] = true
				}
			}

			_, _ = fmt.Fprintf(buf, "\n")

			os.Exit(1)

			return nil
		}
		return err
	}
	return nil
}
