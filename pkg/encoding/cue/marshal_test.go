package cue

import (
	"fmt"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestMarshal(t *testing.T) {
	data, err := Marshal(map[string]any{
		"ok": true,
		"stdout": `1
2
3`,
	})

	testingx.Expect(t, err, testingx.Be[error](nil))
	fmt.Println(string(data))
}
