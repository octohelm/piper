package cueify

import (
	"bytes"
	"encoding"
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
)

type Decl struct {
	PkgPath string
	Name    string
	Doc     []string
	Source  []byte
	Imports map[string]string
	Fields  map[string]*Field
}

type CustomCueType interface {
	CueType() []byte
}

func WithPkgPathReplaceFunc(replace func(pkgPath string) string) OptionFunc {
	return func(s *scanner) {
		s.pkgPathReplace = replace
	}
}

func WithRegister(register func(t reflect.Type)) OptionFunc {
	return func(s *scanner) {
		s.register = register
	}
}

type OptionFunc func(s *scanner)

func FromType(tpe reflect.Type, optFns ...OptionFunc) *Decl {
	for tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}

	s := &scanner{
		pkgPath:    tpe.PkgPath(),
		defs:       map[reflect.Type]bool{},
		fieldInfos: map[reflect.Type]map[string]*Field{},
		imports:    map[string]string{},
		pkgPathReplace: func(pkgPath string) string {
			return pkgPath
		},
		register: func(t reflect.Type) {},
	}

	for _, optFn := range optFns {
		optFn(s)
	}

	c := &Decl{
		Name:    "#" + tpe.Name(),
		PkgPath: s.pkgPathReplace(tpe.PkgPath()),
	}

	c.Source = s.CueDecl(tpe, opt{
		naming: c.Name,
	})

	c.Imports = s.imports
	c.Fields, _ = s.fieldInfos[tpe]

	v := reflect.New(tpe).Interface()

	c.Doc, _ = getRuntimeDoc(v)

	return c
}

type scanner struct {
	pkgPath        string
	defs           map[reflect.Type]bool
	imports        map[string]string
	pkgPathReplace func(pkgPath string) string
	register       func(t reflect.Type)
	fieldInfos     map[reflect.Type]map[string]*Field
}

type opt struct {
	naming string
	embed  string
}

func (s *scanner) Named(name string, pkgPath string) string {
	if pkgPath == s.pkgPath {
		return "#" + name
	}

	replaced := s.pkgPathReplace(pkgPath)
	alias := camelcase.LowerSnakeCase(replaced)
	s.imports[replaced] = alias
	return alias + ".#" + name
}

func (s *scanner) CueDecl(tpe reflect.Type, o opt) []byte {
	if o.naming == "" && tpe.PkgPath() != "" {
		if _, ok := s.defs[tpe]; !ok {
			s.defs[tpe] = true
			s.register(tpe)
		}

		if o.embed != "" {
			return []byte(fmt.Sprintf(`%s & { 
  %s 
}`, s.Named(tpe.Name(), tpe.PkgPath()), o.embed))
		}
		return []byte(s.Named(tpe.Name(), tpe.PkgPath()))
	}

	switch x := reflect.New(tpe).Interface().(type) {
	case CustomCueType:
		return x.CueType()
	case encoding.TextMarshaler:
		return []byte("string")
	case OneOfType:
		types := x.OneOf()
		b := bytes.NewBuffer(nil)

		for i := range types {
			t := reflect.TypeOf(types[i])
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if i > 0 {
				b.WriteString(" | ")
			}
			b.Write(s.CueDecl(t, opt{embed: o.embed}))
		}

		return b.Bytes()
	}

	switch tpe.Kind() {
	case reflect.Ptr:
		return []byte(fmt.Sprintf("%s | null", s.CueDecl(tpe.Elem(), opt{embed: o.embed})))
	case reflect.Map:
		return []byte(fmt.Sprintf("[X=%s]: %s", s.CueDecl(tpe.Key(), opt{embed: o.embed}), s.CueDecl(tpe.Elem(), opt{embed: o.embed})))
	case reflect.Slice:
		if tpe.Elem().Kind() == reflect.Uint8 {
			return []byte("bytes")
		}
		return []byte(fmt.Sprintf("[...%s]", s.CueDecl(tpe.Elem(), opt{embed: o.embed})))
	case reflect.Struct:
		b := bytes.NewBuffer(nil)

		fields := map[string]*Field{}
		defer func() {
			s.fieldInfos[tpe] = fields
		}()

		_, _ = fmt.Fprintf(b, `{
`)

		walkFields(tpe, func(field *Field) {
			fields[field.Name] = field

			t := field.Type

			if field.Inline {
				if t.Kind() == reflect.Map {
					_, _ = fmt.Fprintf(b, `[!~"\\$\\$task"]: %s`, s.CueDecl(t.Elem(), opt{
						embed: field.Embed,
					}))
				}
				return
			}

			if len(field.Doc) > 0 {
				for _, l := range field.Doc {
					_, _ = fmt.Fprintf(b, `// %s
`, l)
				}
			}

			if field.Optional {
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				_, _ = fmt.Fprintf(b, "%s?: ", field.Name)
			} else {
				_, _ = fmt.Fprintf(b, "%s: ", field.Name)
			}

			cueType := s.CueDecl(t, opt{
				embed: field.Embed,
			})

			if len(field.Enum) > 0 {
				for i, e := range field.Enum {
					if i > 0 {
						_, _ = fmt.Fprint(b, " | ")
					}
					_, _ = fmt.Fprintf(b, `%q`, e)
				}
			} else {
				_, _ = fmt.Fprintf(b, "%s", cueType)
			}

			if field.DefaultValue != nil {
				switch string(cueType) {
				case "bytes":
					_, _ = fmt.Fprintf(b, ` | *'%s'`, *field.DefaultValue)
				case "string":
					_, _ = fmt.Fprintf(b, ` | *%q`, *field.DefaultValue)
				default:
					_, _ = fmt.Fprintf(b, ` | *%v`, *field.DefaultValue)
				}
			}

			if field.AsOutput {
				_, _ = fmt.Fprintf(b, " @generated()")
			}

			_, _ = fmt.Fprint(b, "\n")
		})

		if strings.HasSuffix(o.naming, "Interface") {
			_, _ = fmt.Fprintf(b, `
...
`)
		}

		_, _ = fmt.Fprintf(b, `}`)

		return b.Bytes()
	case reflect.Interface:
		return []byte("_")
	default:
		return []byte(tpe.Kind().String())
	}
}

type Field struct {
	Name         string
	Doc          []string
	Embed        string
	Idx          int
	Type         reflect.Type
	AsOutput     bool
	Optional     bool
	Inline       bool
	DefaultValue *string
	Enum         []string
}

func (i *Field) EmptyDefaults() (string, bool) {
	if i.Type.PkgPath() != "" {
		return "", false
	}

	switch i.Type.Kind() {
	case reflect.Slice:
		return "", false
	case reflect.Map:
		return "", false
	case reflect.Interface:
		return "", false
	default:
		return fmt.Sprintf("%v", reflect.New(i.Type).Elem()), true
	}
}

func walkFields(st reflect.Type, each func(info *Field)) {
	if st.Kind() != reflect.Struct {
		return
	}

	v := reflect.New(st).Interface()

	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)

		if !ast.IsExported(f.Name) {
			continue
		}

		info := &Field{}
		info.Idx = i
		info.Name = f.Name
		info.Type = f.Type
		if doc, ok := getRuntimeDoc(v, f.Name); ok {
			info.Doc = doc
		}

		jsonTag, hasJsonTag := f.Tag.Lookup("json")
		if !hasJsonTag {
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				walkFields(f.Type, each)
			}
			continue
		}

		if strings.Contains(jsonTag, ",omitempty") {
			info.Optional = true
		}

		if embed, hasEmbedTag := f.Tag.Lookup("embed"); hasEmbedTag {
			info.Embed = embed
		}

		taskTag, hasOutput := f.Tag.Lookup("output")
		if jsonTag == "-" && !hasOutput {
			continue
		}

		if jsonName := strings.SplitN(jsonTag, ",", 2)[0]; jsonName != "" {
			info.Name = jsonName
		}

		if hasOutput {
			attrs := strings.SplitN(taskTag, ",", 2)

			info.AsOutput = true

			if name := attrs[0]; name != "" {
				info.Name = name
			}
		}

		if defaultValue, ok := f.Tag.Lookup("default"); ok {
			info.DefaultValue = &defaultValue
		}

		if enumValue, ok := f.Tag.Lookup("enum"); ok {
			info.Enum = strings.Split(enumValue, ",")
		}

		if strings.Contains(jsonTag, ",inline") {
			info.Inline = true
			info.Name = ""
		}

		each(info)
	}
}

func getRuntimeDoc(v any, names ...string) ([]string, bool) {
	if c, ok := v.(interface {
		RuntimeDoc(names ...string) ([]string, bool)
	}); ok {
		return c.RuntimeDoc(names...)
	}
	return nil, false
}
