package ocitar

import (
	"encoding"

	"github.com/octohelm/crkit/pkg/artifact/kubepkg/renamer"
)

var _ encoding.TextUnmarshaler = &Rename{}

type Rename struct {
	renamer.Renamer
}

func (r *Rename) CueType() []byte {
	return []byte("string")
}

func (r *Rename) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	re, err := renamer.NewTemplate(string(text))
	if err != nil {
		return err
	}

	*r = Rename{
		Renamer: re,
	}

	return nil
}
