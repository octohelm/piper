package container

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"dagger.io/dagger"
	"github.com/octohelm/piper/pkg/engine/task/client"
)

type Mounter interface {
	MountType() string
	MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error)
}

type Mount struct {
	Mounter `json:"-"`
}

func (m *Mount) UnmarshalJSON(data []byte) error {
	mt := &struct {
		Type string `json:"type"`
	}{}

	if err := json.Unmarshal(data, mt); err != nil {
		return err
	}

	for _, v := range m.OneOf() {
		if i, ok := v.(Mounter); ok {
			if i.MountType() == mt.Type {
				i = reflect.New(reflect.TypeOf(i).Elem()).Interface().(Mounter)

				if err := json.Unmarshal(data, i); err != nil {
					return err
				}

				m.Mounter = i
				return nil
			}
		}
	}

	return nil
}

func (m Mount) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Mounter)
}

func (Mount) OneOf() []any {
	return []any{
		&MountCacheDir{},
		&MountTemp{},
		&MountFs{},
		&MountSecret{},
		&MountFile{},
	}
}

var (
	_ Mounter = &MountCacheDir{}
	_ Mounter = &MountTemp{}
	_ Mounter = &MountFs{}
	_ Mounter = &MountSecret{}
	_ Mounter = &MountFile{}
)

type MountCacheDir struct {
	Type     string   `json:"type" enum:"cache"`
	Dest     string   `json:"dest"`
	Contents CacheDir `json:"contents"`
}

type CacheDir struct {
	ID string `json:"id"`
}

func (MountCacheDir) MountType() string {
	return "cache"
}

func (m MountCacheDir) MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error) {
	return container.WithMountedCache(m.Dest, c.CacheVolume(m.Contents.ID), dagger.ContainerWithMountedCacheOpts{}), nil
}

type MountTemp struct {
	Type     string  `json:"type" enum:"tmp"`
	Dest     string  `json:"dest"`
	Contents TempDir `json:"contents"`
}

type TempDir struct{}

func (t MountTemp) MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error) {
	return container.WithMountedTemp(t.Dest), nil
}

func (MountTemp) MountType() string {
	return "tmp"
}

type MountFs struct {
	Type     string  `json:"type" enum:"fs"`
	Contents Fs      `json:"contents"`
	Dest     string  `json:"dest"`
	Source   *string `json:"source,omitzero"`
}

func (f MountFs) MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error) {
	dir, err := f.Contents.Directory(ctx, c)
	if err != nil {
		return nil, err
	}

	if source := f.Source; source != nil {
		src := *source
		// Same as k8s, when
		//
		//		mountPath: /etc/config/xxxfile
		// 		subPath: xxxfile
		//
		// will mount as file
		if strings.HasSuffix(f.Dest, src) {
			return container.WithMountedFile(f.Dest, dir.File(src)), nil
		}

		// otherwise will mount with sub dir
		dir = dir.Directory(src)
	}

	return container.WithMountedDirectory(f.Dest, dir), nil
}

func (MountFs) MountType() string {
	return "fs"
}

type MountSecret struct {
	Type     string        `json:"type" enum:"secret"`
	Dest     string        `json:"dest"`
	Contents client.Secret `json:"contents"`
}

func (m MountSecret) MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error) {
	if s, ok := Secret(ctx, c, &m.Contents); ok {
		return container.WithMountedSecret(m.Dest, s), nil
	}
	return container, nil
}

func (MountSecret) MountType() string {
	return "secret"
}

type MountFile struct {
	Type        string `json:"type" enum:"file"`
	Dest        string `json:"dest"`
	Contents    string `json:"contents"`
	Permissions int    `json:"mask" default:"0o644"`
}

func (m MountFile) MountTo(ctx context.Context, c *dagger.Client, container *dagger.Container) (*dagger.Container, error) {
	f := c.Container().
		WithNewFile("/tmp", m.Contents, dagger.ContainerWithNewFileOpts{
			Permissions: m.Permissions,
		}).
		File("/tmp")

	return container.WithMountedFile(m.Dest, f), nil
}

func (MountFile) MountType() string {
	return "file"
}
