package cueify

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"cuelang.org/go/cue"
	cueformat "cuelang.org/go/cue/format"

	"github.com/octohelm/x/anyjson"
)

type CueWriterOption func(cw *cueWriter)

func AsDecl(decl string) CueWriterOption {
	return func(cw *cueWriter) {
		cw.decl = decl
	}
}

func WithPkg(pkg string) CueWriterOption {
	return func(cw *cueWriter) {
		cw.pkg = pkg
	}
}

func WithStrictValueMatcher(matcher PathMatcher) CueWriterOption {
	return func(cw *cueWriter) {
		cw.strictValueMatcher = matcher
	}
}

func WithType(v any) CueWriterOption {
	return func(cw *cueWriter) {
		switch x := v.(type) {
		case CustomCueType:
			cw.typed = x.CueType()

			if withImport, ok := v.(CustomCueTypeWithImport); ok {
				importPath, alias := withImport.CueTypeImport()
				if cw.imports == nil {
					cw.imports = map[string]string{}
				}
				cw.imports[importPath] = alias
			}

		case string:
			cw.typed = []byte(x)
		}
	}
}

func WithStaticValue(staticValues map[string]any) CueWriterOption {
	return func(cw *cueWriter) {
		cw.staticValues = staticValues
	}
}

func ValueToCue(value any, optFns ...CueWriterOption) ([]byte, error) {
	w := bytes.NewBuffer(nil)

	cw := &cueWriter{}
	cw.build(optFns...)

	if cw.pkg != "" {
		_, _ = fmt.Fprintf(w, `package %s

`, cw.pkg)
	}

	if len(cw.imports) > 0 {
		for importPath, alias := range cw.imports {
			_, _ = fmt.Fprintf(w, `import %s %q
`, alias, importPath)
		}
	}

	if cw.decl != "" {
		_, _ = fmt.Fprintf(w, `#%s: `, cw.decl)
	}

	if typed := cw.typed; len(typed) > 0 {
		_, _ = fmt.Fprintf(w, `%s &`, string(typed))
	}

	if err := cw.writeToCue(w, value, cue.MakePath()); err != nil {
		return nil, err
	}

	return cueformat.Source(w.Bytes(), cueformat.Simplify())
}

type cueWriter struct {
	pkg                string
	decl               string
	typed              []byte
	imports            map[string]string
	staticValues       map[string]any
	strictValueMatcher PathMatcher
}

func (cw *cueWriter) build(optFns ...CueWriterOption) {
	for _, o := range optFns {
		o(cw)
	}
}

func (cw *cueWriter) writeToCue(w io.Writer, value any, path cue.Path) error {
	v, ok := value.(anyjson.Valuer)
	if !ok {
		obj, err := anyjson.FromValue(value)
		if err != nil {
			return err
		}
		v = obj
	}

	if cw.strictValueMatcher == nil || !cw.strictValueMatcher.Match(path) {
		if _, ok := v.(*anyjson.Object); !ok {
			_, _ = fmt.Fprintf(w, "_ | *")
		}
	}

	if cw.staticValues != nil {
		p := path.String()

		if v, ok := cw.staticValues[p]; ok {
			obj, err := anyjson.FromValue(v)
			if err != nil {
				return err
			}

			raw, err := obj.MarshalJSON()
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, `%s`, raw)
			return nil
		}
	}

	switch x := v.(type) {
	case *anyjson.Object:
		_, _ = fmt.Fprintf(w, `{
`)

		for k, v := range x.KeyValues() {
			_, _ = fmt.Fprintf(w, "%q: ", k)

			if err := cw.writeToCue(w, v, cue.MakePath(slices.Concat(path.Selectors(), cue.ParsePath(k).Selectors())...)); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(w)
		}

		_, _ = fmt.Fprintf(w, `}`)

		return nil
	case *anyjson.Array:
		_, _ = fmt.Fprintf(w, `[
`)

		for i, v := range x.IndexedValues() {
			if err := cw.writeToCue(w, v, cue.MakePath(slices.Concat(path.Selectors(), cue.ParsePath(fmt.Sprintf("[%d]", i)).Selectors())...)); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, `,
`)
		}

		_, _ = fmt.Fprintf(w, `]`)
	case *anyjson.String:
		s := x.Value().(string)

		if strings.Contains(s, "\n") {
			_, _ = fmt.Fprintf(w, `"""
%s
"""`, s)

			return nil
		}

		_, _ = fmt.Fprintf(w, `%q`, s)
	default:
		raw, err := x.MarshalJSON()
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, `%s`, raw)

	}

	return nil
}
