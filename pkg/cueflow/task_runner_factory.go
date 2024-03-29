package cueflow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/octohelm/unifs/pkg/filesystem"
	"github.com/pkg/errors"

	cueformat "cuelang.org/go/cue/format"
	"github.com/octohelm/piper/pkg/cueflow/cueify"
)

func RegisterTask(r TaskImplRegister, task FlowTask) {
	r.Register(task)
}

type TaskImplRegister interface {
	Register(t any)
}

type TaskRunnerFactory interface {
	New(task Task) (TaskRunner, error)
}

type TaskFactory interface {
	TaskRunnerResolver
	TaskImplRegister
	Sources(ctx context.Context) (filesystem.FileSystem, error)
}

func NewTaskFactory(domain string) TaskFactory {
	return &factory{
		domain: domain,
		named:  map[string]*namedType{},
	}
}

type factory struct {
	domain string
	named  map[string]*namedType
}

func (f *factory) Register(t any) {
	tpe := reflect.TypeOf(t)
	for tpe.Kind() == reflect.Ptr {
		tpe = tpe.Elem()
	}
	f.register(tpe)
}

func (f *factory) register(tpe reflect.Type) {
	block := cueify.FromType(tpe,
		cueify.WithPkgPathReplaceFunc(func(pkgPath string) string {
			return fmt.Sprintf("%s/%s", f.domain, filepath.Base(pkgPath))
		}),
		cueify.WithRegister(f.register),
	)

	pt := &namedType{
		tpe:          tpe,
		decl:         block,
		outputFields: map[string][]int{},
	}

	if _, ok := reflect.New(tpe).Interface().(FlowTask); ok {
		pt.flowTask = true
	}

	if _, ok := reflect.New(tpe).Interface().(FlowControl); ok {
		pt.flowControl = true
	}

	for _, info := range pt.decl.Fields {
		if info.AsOutput {
			pt.outputFields[info.Name] = info.Loc
		}
	}

	f.named[pt.FullName()] = pt
}

func (f *factory) ResolveTaskRunner(task Task) (TaskRunner, error) {
	if found, ok := f.named[task.Name()]; ok {
		return found.New(task)
	}
	return nil, fmt.Errorf("unknown task `%s`", task)
}

type source struct {
	pkgName string
	imports map[string]string
	bytes.Buffer
}

func (s *source) WriteDecl(named *namedType) {
	for k, v := range named.decl.Imports {
		s.imports[k] = v
	}

	s.WriteString("\n")

	if named.flowControl {
		_, _ = fmt.Fprintf(s, `%s: $$control: name: %q
`, named.decl.Name, strings.ToLower(strings.Trim(named.decl.Name, "#")))
	} else if named.flowTask {
		_, _ = fmt.Fprintf(s, `%s: $$task: name: %q
`, named.decl.Name, named.FullName())
	}

	_, _ = fmt.Fprintf(s, `%s: %s
`, named.decl.Name, named.decl.Source)
}

func (s *source) Source() ([]byte, error) {
	b := bytes.NewBufferString("package " + s.pkgName)

	if len(s.imports) > 0 {
		_, _ = fmt.Fprintf(b, `

import (
`)

		for e := range SortedIter(context.Background(), s.imports) {

			_, _ = fmt.Fprintf(b, `%s %q
`, e.Value, e.Key)
		}

		_, _ = fmt.Fprintf(b, `)
`)
	}

	_, _ = io.Copy(b, s)

	data, err := cueformat.Source(b.Bytes(), cueformat.Simplify())
	if err != nil {
		return nil, errors.Wrapf(err, `format invalid:

%s`, b.Bytes())
	}
	return data, nil
}

func (f *factory) Sources(ctx context.Context) (filesystem.FileSystem, error) {
	sources := map[string]*source{}

	for nt := range SortedIter(ctx, f.named) {
		s, ok := sources[nt.Value.decl.PkgPath]
		if !ok {
			s = &source{
				pkgName: filepath.Base(nt.Value.decl.PkgPath),
				imports: map[string]string{},
			}
			sources[nt.Value.decl.PkgPath] = s
		}

		s.WriteDecl(nt.Value)
	}

	fs := filesystem.NewMemFS()

	for pathPath, s := range sources {
		code, err := s.Source()
		if err != nil {
			return nil, err
		}

		if err := WriteFile(ctx, fs, path.Join(pathPath, s.pkgName+".cue"), code); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func WriteFile(ctx context.Context, fs filesystem.FileSystem, filename string, data []byte) error {
	if err := filesystem.MkdirAll(ctx, fs, filepath.Dir(filename)); err != nil {
		return err
	}
	file, err := fs.OpenFile(ctx, filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}

type namedType struct {
	tpe          reflect.Type
	flowTask     bool
	flowControl  bool
	outputFields map[string][]int
	decl         *cueify.Decl
}

func (nt *namedType) New(planTask Task) (TaskRunner, error) {
	r := &taskRunner{
		task:            planTask,
		inputTaskRunner: reflect.New(nt.tpe),
		outputFields:    map[string][]int{},
	}

	for f, i := range nt.outputFields {
		r.outputFields[f] = i
	}

	return r, nil
}

func (nt *namedType) FullName() string {
	return fmt.Sprintf("%s.%s", nt.decl.PkgPath, nt.decl.Name)
}

func SortedIter[V any](ctx context.Context, m map[string]V) <-chan *Element[V] {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ch := make(chan *Element[V])

	go func() {
		defer func() {
			close(ch)
		}()

		for _, key := range keys {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- &Element[V]{
					Key:   key,
					Value: m[key],
				}
			}
		}
	}()

	return ch
}

// +gengo:runtimedoc=false
type Element[V any] struct {
	Key   string
	Value V
}
