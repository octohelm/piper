package ocitar

import (
	"encoding"

	"github.com/octohelm/crkit/pkg/kubepkg"
)

var _ encoding.TextUnmarshaler = &Rename{}

type Rename struct {
	kubepkg.Renamer
}

func (r *Rename) CueType() []byte {
	return []byte("string")
}

func (r *Rename) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	renamer, err := kubepkg.NewTemplateRenamer(string(text))
	if err != nil {
		return err
	}

	*r = Rename{
		Renamer: renamer,
	}

	return nil
}
