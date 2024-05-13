package wd

import (
	"os"
	"testing"

	"github.com/k0sproject/rig"
	"github.com/octohelm/piper/pkg/sshutil"
	testingx "github.com/octohelm/x/testing"
	"golang.org/x/net/context"
)

func TestFS(t *testing.T) {
	t.Run("local", func(t *testing.T) {
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

		doWDFSTesting(t, local)
	})

	t.Run("ssh", func(t *testing.T) {
		t.Skip()

		ssh, err := sshutil.Load("~/.ssh/config", "colima")
		testingx.Expect(t, err, testingx.Be[error](nil))

		w, err := Wrap(
			&rig.Connection{SSH: ssh},
			WithDir("/tmp"),
			WithUser("root"),
		)
		testingx.Expect(t, err, testingx.Be[error](nil))

		doWDFSTesting(t, w)
	})
}

func doWDFSTesting(t *testing.T, w WorkDir) {
	t.Run("touch file", func(t *testing.T) {
		f, err := w.OpenFile(context.Background(), "1.txt", os.O_WRONLY|os.O_CREATE, os.ModePerm)
		testingx.Expect(t, err, testingx.Be[error](nil))

		_, err = f.Write([]byte("123"))
		testingx.Expect(t, err, testingx.Be[error](nil))

		f.Close()

		t.Run("ls", func(t *testing.T) {
			files := make([]string, 0)
			err = ListFile(w, ".", func(filename string) error {
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
