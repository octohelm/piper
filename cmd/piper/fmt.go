package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/build/filetypes"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/parser"
	"cuelang.org/go/cue/token"

	"github.com/innoai-tech/infra/pkg/cli"
)

func init() {
	cli.AddTo(App, &Fmt{})
}

type Fmt struct {
	cli.C

	FmtRun
}

type FmtRun struct {
	Inputs []string `arg:""`

	Simplify bool `flag:",omitzero" alias:"s"`
	Files    bool `flag:",omitzero"`
	Check    bool `flag:",omitzero"`
}

func (r *FmtRun) Run(ctx context.Context) error {
	formatOpts := make([]format.Option, 0)

	if r.Simplify {
		formatOpts = append(formatOpts, format.Simplify())
	}

	var foundBadlyFormatted bool

	if r.Files {
		hasDots := slices.ContainsFunc(r.Inputs, func(arg string) bool {
			return strings.Contains(arg, "...")
		})
		if hasDots {
			return errors.New(`cannot use "..." in --files mode`)
		}

		if len(r.Inputs) == 0 {
			r.Inputs = []string{"."}
		}

		processFile := func(path string) error {
			file, err := filetypes.ParseFile(path, filetypes.Input)
			if err != nil {
				return err
			}

			if path == "-" {
				contents, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				file.Source = contents
			}

			wasModified, err := r.formatFile(file, formatOpts)
			if err != nil {
				return err
			}
			if wasModified {
				foundBadlyFormatted = true
			}
			return nil
		}

		for _, arg := range r.Inputs {
			if arg == "-" {
				if err := processFile(arg); err != nil {
					return err
				}
				continue
			}

			arg = filepath.Clean(arg)
			if err := filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					name := d.Name()
					isMod := name == "cue.mod"
					isDot := strings.HasPrefix(name, ".") && name != "." && name != ".."
					if path != arg && (isMod || isDot || strings.HasPrefix(name, "_")) {
						return filepath.SkipDir
					}
					return nil
				}

				if !strings.HasSuffix(path, ".cue") {
					return nil
				}

				return processFile(path)
			}); err != nil {
				return err
			}
		}

		return nil
	}

	builds := r.loadFromArgs(r.Inputs, &load.Config{
		Tests:       true,
		Tools:       true,
		AllCUEFiles: true,
		Package:     "*",
		SkipImports: true,
	})

	if len(builds) == 0 {
		return errors.Newf(token.NoPos, "invalid args")
	}

	for _, inst := range builds {
		if err := inst.Err; err != nil {
			return err
		}
		for _, file := range inst.BuildFiles {
			shouldFormat := inst.User || file.Filename == "-" || filepath.Dir(file.Filename) == inst.Dir
			if !shouldFormat {
				continue
			}

			wasModified, err := r.formatFile(file, formatOpts)
			if err != nil {
				return err
			}
			if wasModified {
				foundBadlyFormatted = true
			}
		}
	}

	if r.Check && foundBadlyFormatted {
		return ErrPrintedError
	}

	return nil
}

var ErrPrintedError = errors.New("terminating because of errors")

func (r *FmtRun) loadFromArgs(args []string, cfg *load.Config) []*build.Instance {
	binst := load.Instances(args, cfg)
	if len(binst) == 0 {
		return nil
	}

	return binst
}

func (r *FmtRun) formatFile(file *build.File, opts []format.Option) (bool, error) {
	src, err := readAll(file.Filename, file.Source)
	if err != nil {
		return false, err
	}

	syntax, err := parser.ParseFile(file.Filename, src, parser.ParseComments)
	if err != nil {
		return false, err
	}

	formatted, err := format.Node(syntax, opts...)
	if err != nil {
		return false, err
	}

	if file.Filename == "-" && !r.Check {
		_, _ = os.Stdout.Write(formatted)
	}

	if bytes.Equal(formatted, src) {
		return false, nil
	}

	path := file.Filename

	switch {
	case r.Check:
		_, _ = fmt.Fprintln(os.Stdout, path)
	case file.Filename == "-":
	default:
		if err := os.WriteFile(file.Filename, formatted, 0666); err != nil {
			return false, err
		}
	}
	return true, nil
}

func readAll(filename string, src any) ([]byte, error) {
	if src != nil {
		switch src := src.(type) {
		case string:
			return []byte(src), nil
		case []byte:
			return src, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if src != nil {
				return src.Bytes(), nil
			}
		case io.Reader:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, src); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		return nil, fmt.Errorf("invalid source type %T", src)
	}
	return os.ReadFile(filename)
}
