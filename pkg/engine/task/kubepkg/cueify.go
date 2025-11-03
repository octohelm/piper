package kubepkg

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/octohelm/cuekit/pkg/cueconvert"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	"github.com/octohelm/cuekit/pkg/cuepath"
	"github.com/octohelm/gengo/pkg/camelcase"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/kubepkg"
	"github.com/octohelm/x/anyjson"

	enginetask "github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	enginetask.Registry.Register(&Cueify{})
}

type Cueify struct {
	task.Task
	// KubePkg spec
	KubePkg KubePkg `json:"kubepkg"`
	// pkg name
	PkgName string `json:"pkgName"`
	// OutDir for cue files
	OutDir wd.WorkDir `json:"outDir"`
}

func (r *Cueify) Do(ctx context.Context) error {
	kpkg := kubepkgv1alpha1.KubePkg(r.KubePkg)

	ret, err := kubepkg.Convert(&kpkg)
	if err != nil {
		return err
	}

	oo, err := kubepkg.Extract(ret)
	if err != nil {
		return err
	}

	return r.OutDir.Do(ctx, func(ctx context.Context, wd pkgwd.WorkDir) error {
		writeFile := func(filename string, data []byte) error {
			f, err := wd.OpenFile(ctx, filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = f.Write(data)
			return err
		}

		for o := range oo {
			filename := fmt.Sprintf("%s.cue", strings.ToLower(o.GetName()))

			switch x := o.(type) {
			case *kubepkgv1alpha1.KubePkg:
				x.Kind = ""
				x.APIVersion = ""

				v, err := anyjson.FromValue(x)
				if err != nil {
					return err
				}

				cleaned := anyjson.Merge(anyjson.Valuer(&anyjson.Object{}), v, anyjson.WithEmptyObjectAsNull())

				m, err := cuepath.CompileGlobMatcher(`"*"."{kind,type}"`)
				if err != nil {
					return err
				}

				data, err := cueconvert.Dump(
					cleaned,
					cueconvert.AsDecl(camelcase.UpperCamelCase(o.GetName())),
					cueconvert.WithPkg(r.PkgName),
					cueconvert.WithType(&KubePkg{}),
					cueconvert.WithStrictValueMatcher(m),
				)
				if err != nil {
					return err
				}

				if err := writeFile(filename, data); err != nil {
					return err
				}
			default:
				gvk := o.GetObjectKind().GroupVersionKind()

				filename = fmt.Sprintf("%s.%s.cue", strings.ToLower(o.GetName()), strings.ToLower(gvk.Kind))

				v, err := anyjson.FromValue(o)
				if err != nil {
					return err
				}

				cleaned := anyjson.Merge(anyjson.Valuer(&anyjson.Object{}), v, anyjson.WithEmptyObjectAsNull())

				data, err := cueconvert.Dump(
					cleaned,
					cueconvert.AsDecl(camelcase.UpperCamelCase(o.GetName()+gvk.Kind)),
					cueconvert.WithPkg(r.PkgName),
					cueconvert.WithStaticValue(map[string]any{
						"apiVersion": gvk.GroupVersion().String(),
						"kind":       gvk.Kind,
					}),
				)
				if err != nil {
					return err
				}
				if err := writeFile(filename, data); err != nil {
					return err
				}
			}
		}
		return nil
	}, pkgwd.WithDir(r.PkgName))
}
