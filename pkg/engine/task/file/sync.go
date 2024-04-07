package file

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/octohelm/piper/pkg/otel"

	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/chunk"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Sync{})
}

// Sync file to contents
type Sync struct {
	task.Task
	// source file
	SrcFile File `json:"srcFile"`
	// sync option
	With SyncOption `json:"with,omitempty"`
	// dest fie
	OutFile File `json:"outFile"`
	// synced file same as dest
	File File `json:"-" output:"file"`
}

type SyncOption struct {
	// once maxConcurrent larger than 1,
	// file will split to chunk for partially read and write when syncing
	MaxConcurrent int `json:"maxConcurrent" default:"16"`
}

func (t *Sync) Do(ctx context.Context) error {
	return t.SrcFile.WorkDir.Do(ctx, func(ctx context.Context, src wd.WorkDir) error {
		srcFileInfo, err := src.Stat(ctx, t.SrcFile.Filename)
		if err != nil {
			return errors.Wrapf(err, "%s: get digest failed", src)
		}
		srcDgst, err := src.Digest(ctx, t.SrcFile.Filename)
		if err != nil {
			return errors.Wrapf(err, "%s: get digest failed", src)
		}

		return t.OutFile.WorkDir.Do(ctx, func(ctx context.Context, dst wd.WorkDir) (err error) {
			dstDgst, _ := dst.Digest(ctx, t.OutFile.Filename)

			defer func() {
				if err != nil {
					// when err should remove file
					_ = dst.RemoveAll(ctx, t.OutFile.Filename)
				} else {
					t.File = t.OutFile
					t.Done(nil)
				}
			}()

			// exists and not changed, skip
			if dstDgst == srcDgst {
				return nil
			}

			dir := filepath.Dir(t.OutFile.Filename)
			if !(dir == "" || dir == "/") {
				if err := filesystem.MkdirAll(ctx, dst, dir); err != nil {
					return errors.Wrapf(err, "%s %s: mkdir failed", dst, dir)
				}
			}

			total := srcFileInfo.Size()

			if err := t.truncateDst(ctx, dst, total); err != nil {
				return err
			}

			pw := cueflow.NewProcessWriter(total)
			_, l := logr.FromContext(ctx).Start(ctx, "syncing", slog.Int64(otel.LogAttrProgressTotal, total))
			defer l.End()

			go func() {
				for p := range pw.Process(ctx) {
					l.WithValues(slog.Int64(otel.LogAttrProgressCurrent, p.Current)).Info("")
				}
			}()

			w := chunk.NewWorker(
				chunk.FileSize(total),
				chunk.WithMaxConcurrent(t.With.MaxConcurrent),
			)

			w.Do(func(c chunk.Chunk) error {
				return t.syncN(ctx, src, dst, int64(c.Size), int64(c.Offset), pw)
			})

			if err := w.Wait(); err != nil {
				return err
			}

			dstDgst, err = dst.Digest(ctx, t.OutFile.Filename)
			if err != nil {
				return err
			}

			if dstDgst != srcDgst {
				return errors.Errorf("sync failed, expected %s, but got %s", srcDgst, dstDgst)
			}

			return
		})
	})
}

func (t *Sync) truncateDst(ctx context.Context, dst wd.WorkDir, total int64) error {
	dstFile, err := dst.OpenFile(ctx, t.OutFile.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s %s: open dest file failed", dst, t.OutFile.Filename)
	}
	defer dstFile.Close()

	if f, ok := dstFile.(CanTruncate); ok {
		if err := f.Truncate(total); err != nil {
			return err
		}
		return nil
	}

	return errors.New("Truncate is not supported")
}

func (t *Sync) syncN(ctx context.Context, src wd.WorkDir, dst wd.WorkDir, size int64, offset int64, alt io.Writer) error {
	srcFile, err := src.OpenFile(ctx, t.SrcFile.Filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s  %s: open source file failed", src, t.SrcFile.Filename)
	}
	defer srcFile.Close()

	if _, err := srcFile.Seek(offset, io.SeekCurrent); err != nil {
		return err
	}

	dstFile, err := dst.OpenFile(ctx, t.OutFile.Filename, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s %s: open dest file failed", src, t.OutFile.Filename)
	}
	defer dstFile.Close()

	if _, err := dstFile.Seek(offset, io.SeekCurrent); err != nil {
		return err
	}

	if _, err := io.CopyN(dstFile, io.TeeReader(srcFile, alt), size); err != nil {
		return err
	}

	return nil
}

type CanTruncate interface {
	Truncate(size int64) error
}
