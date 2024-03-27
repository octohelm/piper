package engine

import (
	"bytes"
	"fmt"

	cueerrors "cuelang.org/go/cue/errors"
	cuetoken "cuelang.org/go/cue/token"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Pipeline struct {
	Action []string `arg:""`
	// plan root file
	// and the dir of the root file will be the cwd for all cue files
	Project string `flag:",omitempty" alias:"p"`

	// cache dir root
	// for cache files
	CacheDir string `flag:",omitempty"`
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
			buf := bytes.NewBuffer(nil)
			_, _ = fmt.Fprintf(buf, "%s\n", err)

			records := map[cuetoken.Pos]bool{}
			for _, e := range errList {
				if _, ok := records[e.Position()]; !ok {
					cueerrors.Print(buf, e, nil)
					records[e.Position()] = true
				}
			}
			return errors.New(buf.String())
		}
		return err
	}
	return nil
}
