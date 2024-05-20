package cueify

import (
	"testing"

	"cuelang.org/go/cue"
	testingx "github.com/octohelm/x/testing"
)

func TestPathMatcher(t *testing.T) {
	m := CreatePathMatcher(`"*"."{kind,type}"`)

	testingx.Expect(t, m.Match(cue.ParsePath("x.kind")), testingx.BeTrue())
	testingx.Expect(t, m.Match(cue.ParsePath("a.b.c.kind")), testingx.BeTrue())
	testingx.Expect(t, m.Match(cue.ParsePath("a.b")), testingx.BeFalse())
}
