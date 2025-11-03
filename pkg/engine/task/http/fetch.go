package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/octohelm/x/logr"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	"github.com/octohelm/piper/pkg/otel"
	"github.com/octohelm/piper/pkg/progress"
)

func init() {
	enginetask.Registry.Register(&Fetch{})
}

// Fetch http resource to local cache
type Fetch struct {
	task.Task

	// http request url
	Url string `json:"url"`
	// hit by response header
	HitBy string `json:"hitBy,omitzero" default:"etag"`

	// fetched file
	File file.File `json:"-" output:"file"`
}

func (r *Fetch) Do(ctx context.Context) (e error) {
	cwd, err := enginetask.ClientContext.From(ctx).CacheDir(ctx, "http")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Get(r.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	newHitValue := resp.Header.Get(r.HitBy)
	key := base64.RawStdEncoding.EncodeToString([]byte(r.Url))

	currentHitValue, err := readHitValue(ctx, cwd, key)
	if err != nil {
		return err
	}

	if newHitValue == "" || newHitValue != currentHitValue {
		size := resp.ContentLength

		var reader io.Reader = resp.Body

		if size > 0 {
			pw := progress.NewWriter(size)

			_, l := logr.FromContext(ctx).Start(ctx, "downloading", slog.Int64(otel.LogAttrProgressTotal, size))
			defer l.End()

			go func() {
				for p := range pw.Process(ctx) {
					l.WithValues(slog.Int64(otel.LogAttrProgressCurrent, p.Current)).Info("")
				}
			}()

			reader = io.TeeReader(reader, pw)
		}

		if err := copyToCache(ctx, cwd, key, newHitValue, reader); err != nil {
			return err
		}
	}

	return r.File.Sync(ctx, cwd, path.Join(key, "data"))
}

func readHitValue(ctx context.Context, fs filesystem.FileSystem, key string) (string, error) {
	f, err := fs.OpenFile(ctx, path.Join(key, "hit"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func copyToCache(ctx context.Context, fs filesystem.FileSystem, key string, hitValue string, dataReader io.Reader) error {
	if err := filesystem.MkdirAll(ctx, fs, key); err != nil {
		return err
	}

	tmpData := path.Join(key, "data.tmp")

	if err := writeTo(ctx, fs, tmpData, dataReader); err != nil {
		return err
	}
	if err := fs.Rename(ctx, tmpData, path.Join(key, "data")); err != nil {
		return err
	}
	if err := writeTo(ctx, fs, path.Join(key, "hit"), bytes.NewBufferString(hitValue)); err != nil {
		return err
	}

	return nil
}

func writeTo(ctx context.Context, fs filesystem.FileSystem, filename string, r io.Reader) error {
	f, err := fs.OpenFile(ctx, filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}
