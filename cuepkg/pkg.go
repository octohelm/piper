package cuepkg

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/octohelm/cuekit/pkg/cueflow/runner"
	"github.com/octohelm/cuekit/pkg/mod/modmem"
	"github.com/octohelm/unifs/pkg/filesystem"

	"github.com/octohelm/piper/pkg/engine/task"
)

func RegisterAsMemModule() error {
	base := "piper.octohelm.tech"

	m, err := modmem.NewModule(base, "v0.0.0-builtin", func(ctx context.Context, fsDest filesystem.FileSystem) error {
		fsSrc, err := runner.Source(ctx, task.Registry)
		if err != nil {
			return err
		}

		return listFile(ctx, fsSrc, base, func(filename string) error {
			src, err := fsSrc.OpenFile(ctx, filepath.Join(base, filename), os.O_RDONLY, os.ModePerm)
			if err != nil {
				return fmt.Errorf("open source file failed: %w", err)
			}
			defer src.Close()

			if err := filesystem.MkdirAll(ctx, fsDest, filepath.Dir(filename)); err != nil {
				return err
			}
			dest, err := fsDest.OpenFile(ctx, filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			if err != nil {
				return fmt.Errorf("open dest file failed: %w", err)
			}
			defer dest.Close()

			if _, err := io.Copy(dest, src); err != nil {
				return err
			}
			return nil
		})
	})
	if err != nil {
		return err
	}

	modmem.DefaultRegistry.Register(m)

	return nil
}

func listFile(ctx context.Context, fsys filesystem.FileSystem, root string, each func(filename string) error) error {
	return filesystem.WalkDir(ctx, fsys, root, func(path string, d fs.DirEntry, err error) error {
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
