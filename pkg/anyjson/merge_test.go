package anyjson

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestMerge(t *testing.T) {
	base := From(Map{
		"int": 1,
		"str": "string",
	}).(*Object)

	t.Run("normal merge", func(t *testing.T) {
		patch := From(Map{
			"str":    "changed",
			"extra":  true,
			"ignore": nil,
		}).(*Object)

		merged := Merge(base, patch)

		testingx.Expect(t, merged.Value(), testingx.Equal[any](Map{
			"int":   1.0,
			"str":   "changed",
			"extra": true,
		}))
	})

	t.Run("nil remover", func(t *testing.T) {
		patch := From(Map{
			"str": nil,
		}).(*Object)

		merged := Merge(base, patch, WithNullOp(NullAsRemover))

		testingx.Expect(t, merged.Value(), testingx.Equal[any](Map{
			"int": 1.0,
		}))
	})

	t.Run("array object merge", func(t *testing.T) {
		base := From(List{
			Map{
				"name":  "a",
				"value": "x",
			},
			Map{
				"name":  "b",
				"value": "x",
			},
		})

		patch := From(List{
			Map{
				"name":  "a",
				"value": "patched",
			},
			Map{
				"name":  "c",
				"value": "new",
			},
		})

		merged := Merge(base, patch, WithArrayMergeKey("name"))

		testingx.Expect(t, merged.Value(), testingx.Equal[any](List{
			Map{
				"name":  "a",
				"value": "patched",
			},
			Map{
				"name":  "b",
				"value": "x",
			},
			Map{
				"name":  "c",
				"value": "new",
			},
		}))
	})
}
