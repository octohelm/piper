package container

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/octohelm/piper/pkg/cueflow"
	piperdagger "github.com/octohelm/piper/pkg/dagger"
	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
	pkgwd "github.com/octohelm/piper/pkg/wd"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Export{})
}

type Export struct {
	task.Task

	Input Container `json:"input"`

	// oci annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	OutFile file.File `json:"outFile"`

	File file.File `json:"-" output:"file"`
}

func (x *Export) Do(ctx context.Context) error {
	return x.Input.Select(ctx).Do(ctx, func(ctx context.Context, c *piperdagger.Client) error {
		w, err := x.OutFile.WorkDir.Get(ctx)
		if err != nil {
			return err
		}

		base, err := pkgwd.RealPath(w)
		if err != nil {
			return fmt.Errorf("%T: only support cwd in local host", x)
		}

		cc, err := x.Input.Container(ctx, c)
		if err != nil {
			return err
		}

		if len(x.Annotations) > 0 {
			for k, v := range x.Annotations {
				cc = cc.WithAnnotation(k, v)
			}
		}

		output, err := cc.Export(ctx, filepath.Join(base, x.OutFile.Filename), dagger.ContainerExportOpts{
			MediaTypes: dagger.Ocimediatypes,
		})
		if err != nil {
			return err
		}

		//if len(x.Annotations) > 0 {
		//	if err := ocitar.Replace(output, func(hdr *tar.Header, r io.Reader) (io.Reader, error) {
		//		if hdr.Name == "index.json" {
		//			index := &specv1.Index{}
		//			if err := json.NewDecoder(r).Decode(index); err != nil {
		//				return nil, err
		//			}
		//
		//			for i, m := range index.Manifests {
		//				if m.Annotations == nil {
		//					m.Annotations = make(map[string]string)
		//				}
		//
		//				for k, v := range x.Annotations {
		//					m.Annotations[k] = v
		//				}
		//
		//				index.Manifests[i] = m
		//			}
		//
		//			raw, err := json.Marshal(index)
		//			if err != nil {
		//				return nil, err
		//			}
		//
		//			hdr.Size = int64(len(raw))
		//
		//			return bytes.NewBuffer(raw), nil
		//		}
		//
		//		return r, nil
		//	}); err != nil {
		//		return err
		//	}
		//}

		if len(output) > 0 {
			return x.File.SyncWith(ctx, x.OutFile)
		}

		return errors.New("export failed")
	})
}
