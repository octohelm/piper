package file

import (
	"context"
	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/chunk"
	"github.com/pkg/errors"
	"io"
	"log/slog"
	"os"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Sync{})
}

// Sync file to contents
type Sync struct {
	task.Task
	// source file
	Source File `json:"source"`
	// sync option
	With SyncOption `json:"with"`
	// dest fie
	Dest File `json:"dest"`

	WrittenFileResult `json:"-" output:"result"`
}

type SyncOption struct {
	// once maxConcurrent larger than 1,
	// file will split to chunk for partially read and write when syncing
	MaxConcurrent int `json:"maxConcurrent" default:"16"`
}

var _ cueflow.WithScopeName = &Sync{}

func (w *Sync) ScopeName(ctx context.Context) (string, error) {
	return w.Dest.Wd.ScopeName(ctx)
}

func (t *Sync) Do(ctx context.Context) error {
	return t.Source.Wd.Do(ctx, func(ctx context.Context, src wd.WorkDir) error {
		srcFileInfo, err := src.Stat(ctx, t.Source.Filename)
		if err != nil {
			return errors.Wrapf(err, "%s: get digest failed", src)
		}
		srcDgst, err := src.Digest(ctx, t.Source.Filename)
		if err != nil {
			return errors.Wrapf(err, "%s: get digest failed", src)
		}

		return t.Dest.Wd.Do(ctx, func(ctx context.Context, dst wd.WorkDir) (err error) {
			dstDgst, _ := dst.Digest(ctx, t.Dest.Filename)

			defer func() {
				if err != nil {
					// when err should remove file
					_ = dst.RemoveAll(ctx, t.Dest.Filename)
				} else {
					t.WrittenFileResult.Ok = true
					t.WrittenFileResult.File = t.Dest
				}
			}()

			// exists and not changed, skip
			if dstDgst == srcDgst {
				return nil
			}

			total := srcFileInfo.Size()

			if err := t.truncateDst(ctx, dst, total); err != nil {
				return err
			}

			pw := cueflow.NewProcessWriter(total)
			_, l := logr.FromContext(ctx).Start(ctx, "syncing", slog.Int64(cueflow.LogAttrProgressTotal, total))
			defer l.End()

			go func() {
				for p := range pw.Process(ctx) {
					l.WithValues(slog.Int64(cueflow.LogAttrProgressCurrent, p.Current)).Info("")
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

			dstDgst, err = dst.Digest(ctx, t.Dest.Filename)
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
	dstFile, err := dst.OpenFile(ctx, t.Dest.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s: open file failed", dst)
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
	srcFile, err := src.OpenFile(ctx, t.Source.Filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s: open source file failed", src)
	}
	defer srcFile.Close()

	if _, err := srcFile.Seek(offset, io.SeekCurrent); err != nil {
		return err
	}

	dstFile, err := dst.OpenFile(ctx, t.Dest.Filename, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "%s: open dest file failed", src)
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
