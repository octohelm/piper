package internal

import (
	"strings"

	"cuelang.org/go/cue"
	"github.com/octohelm/piper/pkg/wd"
)

func FormatAsJSONPath(p cue.Path) string {
	b := &strings.Builder{}

	for i, s := range p.Selectors() {
		if i > 0 {
			b.WriteRune('/')
		}

		if s.Type() == cue.StringLabel {
			if strings.Contains(s.String(), "/") {
				b.WriteString(s.String())
				continue
			}
			b.WriteString(s.Unquoted())
			continue
		}
		b.WriteString(s.String())
	}

	return b.String()
}

func PlatformScoped(p cue.Path) (*wd.Platform, bool) {
	for _, s := range p.Selectors() {
		if s.Type() == cue.StringLabel {
			platform, err := wd.ParsePlatform(s.Unquoted())
			if err == nil && platform.Architecture != "" && supportedPlatforms[platform.OS] {
				return platform, true
			}
		}
	}
	return nil, false
}

var supportedPlatforms = map[string]bool{
	"linux":   true,
	"windows": true,
	"darwin":  true,
}
