package kubepkg

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/octohelm/x/anyjson"

	"github.com/octohelm/gengo/pkg/camelcase"
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
	"github.com/octohelm/kubepkgspec/pkg/kubepkg"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/cueflow/cueify"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/wd"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Cueify{})
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

				data, err := cueify.ValueToCue(
					cleaned,
					cueify.AsDecl(camelcase.UpperCamelCase(o.GetName())),
					cueify.WithPkg(r.PkgName),
					cueify.WithType(&KubePkg{}),
					cueify.WithStrictValueMatcher(cueify.CreatePathMatcher(`"*"."{kind,type}"`)),
				)
				if err != nil {
					return err
				}

				if err := writeFile(filename, data); err != nil {
					return err
				}
			default:
				gvk := o.GetObjectKind().GroupVersionKind()

				v, err := anyjson.FromValue(o)
				if err != nil {
					return err
				}

				cleaned := anyjson.Merge(anyjson.Valuer(&anyjson.Object{}), v, anyjson.WithEmptyObjectAsNull())

				data, err := cueify.ValueToCue(
					cleaned,
					cueify.AsDecl(camelcase.UpperCamelCase(o.GetName())),
					cueify.WithPkg(r.PkgName),
					cueify.WithStaticValue(map[string]any{
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
