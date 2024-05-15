package kubepkg

import (
	kubepkgv1alpha1 "github.com/octohelm/kubepkgspec/pkg/apis/kubepkg/v1alpha1"
)

type KubePkg kubepkgv1alpha1.KubePkg

func (KubePkg) CueTypeImport() (importPath string, alias string) {
	return "github.com/octohelm/kubepkgspec/cuepkg/kubepkg", "kubepkg"
}

func (KubePkg) CueType() []byte {
	return []byte("kubepkg.#KubePkg")
}
