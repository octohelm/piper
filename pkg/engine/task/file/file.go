package file

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

type File struct {
	// current work dir
	Cwd wd.WorkDir `json:"cwd"`
	// filename related from current work dir
	Filename string `json:"filename"`
}

func (f *File) SetBy(ctx context.Context, cwd pkgwd.WorkDir, filename string) {
	f.Cwd.SetBy(ctx, cwd)
	f.Filename = filename
}

type StringOrFile struct {
	File   *File
	String string
}

func (StringOrFile) OneOf() []any {
	return []any{
		"",
		&File{},
	}
}

func (s *StringOrFile) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		b := ""
		if err := json.Unmarshal(data, &b); err != nil {
			return err
		}
		*s = StringOrFile{
			String: b,
		}
		return nil
	}

	*s = StringOrFile{
		File: &File{},
	}
	return json.Unmarshal(data, s.File)
}

func (s StringOrFile) MarshalJSON() ([]byte, error) {
	if s.File != nil {
		return json.Marshal(s.File)
	}
	return json.Marshal(s.String)
}

func (s *StringOrFile) Size(ctx context.Context) (int64, error) {
	if s.File != nil {
		cwd, err := s.File.Cwd.Get(ctx)
		if err != nil {
			return -1, err
		}
		info, err := cwd.Stat(ctx, s.File.Filename)
		if err != nil {
			return -1, err
		}
		return info.Size(), nil
	}
	return int64(len(s.String)), nil
}

func (s *StringOrFile) Open(ctx context.Context) (io.ReadCloser, error) {
	if s.File != nil {
		cwd, err := s.File.Cwd.Get(ctx)
		if err != nil {
			return nil, err
		}
		return cwd.OpenFile(ctx, s.File.Filename, os.O_RDONLY, os.ModePerm)
	}

	if len(s.String) == 0 {
		return nil, nil
	}

	return io.NopCloser(bytes.NewBufferString(s.String)), nil
}
