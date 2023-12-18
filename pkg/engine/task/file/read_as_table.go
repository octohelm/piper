package file

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/octohelm/piper/pkg/cueflow"

	"github.com/octohelm/piper/pkg/engine/task"
	taskwd "github.com/octohelm/piper/pkg/engine/task/wd"
	"github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &ReadAsTable{})
}

// ReadAsTable file read as table
type ReadAsTable struct {
	task.Task

	taskwd.CurrentWorkDir
	// filename
	Filename string `json:"filename"`

	// options
	With ReadAsTableOption `json:"with,omitempty"`

	// file contents
	ReadAsTableResult `json:"-" output:"result"`
}

type ReadAsTableOption struct {
	// strict column num
	StrictColNum int `json:"strictColNum,omitempty"`
}

type ReadAsTableResult struct {
	cueflow.Result
	// file contents
	Data [][]string `json:"data"`
}

func (t *ReadAsTableResult) ResultValue() any {
	return t
}

func (t *ReadAsTable) Do(ctx context.Context) error {
	return t.Cwd.Do(ctx, func(ctx context.Context, cwd wd.WorkDir) (err error) {
		defer t.Done(err)

		f, err := cwd.OpenFile(ctx, t.Filename, os.O_RDONLY, os.ModePerm)
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
