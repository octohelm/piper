package cuepkg

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/pkg/errors"

	"github.com/octohelm/cuemod/pkg/cuemod/stdlib"
	"github.com/octohelm/unifs/pkg/filesystem"
)

var (
	PiperModule = "piper.octohelm.tech"
)

func RegistryCueStdlibs() error {
	source, err := task.Factory.Sources(context.Background())
	if err != nil {
		return err
	}

	module, err := createModule(filesystem.AsReadDirFS(source))
	if err != nil {
		return err
	}

	// ugly lock embed version
	if err := registerStdlib(filesystem.AsReadDirFS(module), "v0.0.0", PiperModule); err != nil {
		return err
	}

	return nil
}

func registerStdlib(fs fs.ReadDirFS, ver string, modules ...string) error {
	stdlib.Register(fs, ver, modules...)
	return nil
}

func createModule(otherFs ...fs.ReadDirFS) (filesystem.FileSystem, error) {
	mfs := filesystem.NewMemFS()

	ctx := context.Background()

	for _, f := range otherFs {
		if err := listFile(f, ".", func(filename string) error {
			src, err := f.Open(filename)
			if err != nil {
				return errors.Wrap(err, "open source file failed")
			}
			defer src.Close()

			if err := filesystem.MkdirAll(ctx, mfs, filepath.Dir(filename)); err != nil {
				return err
			}
			dest, err := mfs.OpenFile(ctx, filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "open dest file failed")
			}
			defer dest.Close()

			if _, err := io.Copy(dest, src); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return mfs, nil
}

func listFile(f fs.ReadDirFS, root string, each func(filename string) error) error {
	return fs.WalkDir(f, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := path
		if root != "" && root != "." {
			rel, _ = filepath.Rel(root, path)
		}
		return each(rel)
	})
}
