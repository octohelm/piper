package engine

import (
	"bytes"
	cueerrors "cuelang.org/go/cue/errors"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type Pipeline struct {
	Action []string `arg:""`
	Plan   string   `flag:",omitempty" alias:"p"`
}

func (c *Pipeline) SetDefaults() {
	if c.Plan == "" {
		c.Plan = "./piper.cue"
	}
}

func (c *Pipeline) Run(ctx context.Context) error {
	p, err := New(ctx, WithPlan(c.Plan))
	if err != nil {
		return err
	}

	if err := p.Run(ctx, c.Action...); err != nil {
		if errList := cueerrors.Errors(err); len(errList) > 0 {
			buf := bytes.NewBuffer(nil)
			for i := range errList {
				cueerrors.Print(buf, errList[i], nil)
			}
			return errors.New(buf.String())
		}
		return err
	}

	return nil
}
