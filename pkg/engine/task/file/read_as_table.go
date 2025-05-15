package file

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&ReadAsTable{})
}

// ReadAsTable file read as table
type ReadAsTable struct {
	task.Task
	// file
	File File `json:"file"`
	// options
	With ReadAsTableOption `json:"with,omitzero"`
	// file contents
	Data [][]string `json:"-" output:"data"`
}

type ReadAsTableOption struct {
	// strict column num
	StrictColNum int `json:"strictColNum,omitzero"`
}

func (t *ReadAsTable) Do(ctx context.Context) error {
	return t.File.WorkDir.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		f, err := cwd.OpenFile(ctx, t.File.Filename, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer f.Close()

		tr := &tableReader{
			opt: t.With,
		}

		t.Data = tr.ReadAsTable(f)

		return nil
	})
}

type tableReader struct {
	opt ReadAsTableOption
}

func (tr *tableReader) ReadAsTable(r io.Reader) [][]string {
	rows := make([][]string, 0)

	s := bufio.NewScanner(r)

	for s.Scan() {
		line := s.Text()
		// skip empty and comments
		if line == "" || line[0] == '#' {
			continue
		}
		rows = append(rows, tr.normalize(strings.Fields(line)))
	}

	return rows
}

func (tr *tableReader) normalize(cols []string) (finals []string) {
	if tr.opt.StrictColNum > 0 {
		finals = make([]string, tr.opt.StrictColNum)
	} else {
		finals = make([]string, len(cols))
	}

	for i, col := range cols {
		finals[i] = col
	}

	return
}
