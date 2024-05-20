package cueify

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
	"golang.org/x/tools/txtar"
)

func TestValueToCue(t *testing.T) {
	a := &txtar.Archive{}

	{
		v, err := ValueToCue(map[string]any{
			"x": 1,
		}, WithStrictValueMatcher(CreatePathMatcher("x")))
		testingx.Expect(t, err, testingx.BeNil[error]())

		a.Files = append(a.Files, txtar.File{
			Name: "simple.cue",
			Data: v,
		})
	}

	{
		v2, err := ValueToCue(map[string]any{
			"kind": nil,
			"x":    1,
		}, WithStaticValue(map[string]any{
			"kind": "X",
		}))
		testingx.Expect(t, err, testingx.BeNil[error]())

		a.Files = append(a.Files, txtar.File{
			Name: "with_static.cue",
			Data: v2,
		})
	}

	{
		v3, err := ValueToCue(map[string]any{
			"x": 1,
		}, WithType("X"))
		testingx.Expect(t, err, testingx.BeNil[error]())

		a.Files = append(a.Files, txtar.File{
			Name: "with_typed.cue",
			Data: v3,
		})
	}

	{
		v3, err := ValueToCue(map[string]any{
			"x": 1,
		}, AsDecl("X"))
		testingx.Expect(t, err, testingx.BeNil[error]())

		a.Files = append(a.Files, txtar.File{
			Name: "as_decl.cue",
			Data: v3,
		})
	}

	testingx.Expect(t, a, testingx.MatchSnapshot("value_to_cue"))
}
