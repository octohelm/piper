package wd

import (
	"github.com/k0sproject/rig"
	testingx "github.com/octohelm/x/testing"
	"golang.org/x/net/context"
	"os"
	"testing"
)

func TestFS(t *testing.T) {
	dir := t.TempDir()
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	local, err := Wrap(
		&rig.Connection{
			Localhost: &rig.Localhost{
				Enabled: true,
			},
		},
		WithDir(dir),
		WithUser("root"),
	)
	testingx.Expect(t, err, testingx.Be[error](nil))

	t.Run("touch file", func(t *testing.T) {
		f, err := local.OpenFile(context.Background(), "1.txt", os.O_RDWR|os.O_CREATE, os.ModePerm)
		testingx.Expect(t, err, testingx.Be[error](nil))
		f.Close()

		t.Run("ls", func(t *testing.T) {
			files := make([]string, 0)

			err = ListFile(local, ".", func(filename string) error {
				files = append(files, filename)
				return nil
			})
			testingx.Expect(t, err, testingx.Be[error](nil))
			testingx.Expect(t, files, testingx.Equal([]string{
				"1.txt",
			}))
		})
	})
}
