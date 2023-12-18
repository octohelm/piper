package cuepkg

import (
	"fmt"
	"github.com/octohelm/piper/pkg/wd"
	"testing"
)

func TestCuePkg(t *testing.T) {
	cuepkgs, err := createWagonModule(daggerPortalModules)
	if err != nil {
		t.Fatal(err)
	}

	_ = wd.ListFile(cuepkgs, "", func(filename string) error {
		fmt.Println(filename)
		return nil
	})
}
