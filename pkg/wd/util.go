package wd

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/octohelm/unifs/pkg/filesystem"
)

func ListFile(f filesystem.FileSystem, root string, each func(filename string) error) error {
	return filesystem.WalkDir(context.Background(), f, root, func(path string, d fs.DirEntry, err error) error {
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

type Dir string

func (d Dir) String() string {
	return string(d)
}

func (d Dir) With(dir string) Dir {
	if dir != "" {
		if d == "" || strings.HasPrefix(dir, "/") {
			return Dir(dir)
		} else {
			prefix := string(d)
			if prefix == "" {
				prefix = "/"
			}
			return Dir(filepath.Join(prefix, dir))
		}
	}
	return d
}
