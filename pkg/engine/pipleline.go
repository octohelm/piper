package engine

import (
	"bytes"
	context0 "context"
	"fmt"
	"os"

	cueerrors "cuelang.org/go/cue/errors"
	cuetoken "cuelang.org/go/cue/token"
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

func (pipeline *Pipeline) Run(ctx context0.Context) error {
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
			buf := bytes.NewBuffer(nil)
			_, _ = fmt.Fprintf(buf, "\n")

			printed := map[cuetoken.Pos]bool{}
			for _, e := range errList {
				if _, ok := printed[e.Position()]; !ok {
					cueerrors.Print(buf, e, nil)
					printed[e.Position()] = true
				}
			}

			_, _ = fmt.Fprintf(buf, "\n")

			fmt.Println(buf.String())

			os.Exit(1)

			return nil
		}
		return err
	}
	return nil
}
