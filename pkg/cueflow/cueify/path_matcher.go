package cueify

import (
	"cuelang.org/go/cue"
	"github.com/gobwas/glob"
	"github.com/octohelm/piper/pkg/cueflow/internal"
)

type PathMatcher interface {
	Match(p cue.Path) bool
}

func CreatePathMatcher(rules ...string) PathMatcher {
	m := &matcher{}

	for _, x := range rules {
		m.rules = append(m.rules, glob.MustCompile(internal.FormatAsJSONPath(cue.ParsePath(x))))
	}

	return m
}

type matcher struct {
	rules []glob.Glob
}

func (m *matcher) Match(p cue.Path) bool {
	if m.rules == nil {
		return false
	}

	targetPath := internal.FormatAsJSONPath(p)

	for _, r := range m.rules {
		if r.Match(targetPath) {
			return true
		}
	}
	return false
}
