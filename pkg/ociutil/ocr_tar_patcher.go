package ociutil

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ManifestPatcher struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

type OciTarPatcher struct {
	Manifests []ManifestPatcher `json:"manifests"`
}

func (p *OciTarPatcher) PatchTo(w io.Writer, r io.Reader) error {
	if p.Manifests == nil {
		_, err := io.Copy(w, r)
		return err
	}

	tr := tar.NewReader(r)
	tw := tar.NewWriter(w)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if hdr.Name == "index.json" {
			index := &v1.Index{}
			if err := json.NewDecoder(tr).Decode(index); err != nil {
				return err
			}

			for i := range index.Manifests {
				m := index.Manifests[i]

				if i < len(p.Manifests) {
					if m.Annotations == nil {
						m.Annotations = map[string]string{}
					}

					for k, v := range p.Manifests[i].Annotations {
						m.Annotations[k] = v
					}
				}

				index.Manifests[i] = m
			}

			b := bytes.NewBuffer(nil)
			if err := json.NewEncoder(b).Encode(index); err != nil {
				return err
			}

			// change size
			hdr.Size = int64(b.Len())
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := tw.Write(b.Bytes()); err != nil {
				return err
			}

			continue
		}

		// normal copy
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}

	return tw.Close()
}
