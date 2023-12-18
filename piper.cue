package main

import (
	"path"
	"strconv"

	"piper.octohelm.tech/core"
)

hosts: {
	local: core.#RootfsFromLocal & {

	}
}

ver: core.#RevInfo & {

}

actions: {
	go: {
		main: "./cmd/piper"
		os: ["darwin", "linux", "windows"]
		arch: ["amd64", "arm64"]
		bin: string | *path.Base(main)

		build: {
			for _os in os for _arch in arch {
				"\(_os)/\(_arch)": {
					_build: core.#Exec & {
						cwd: hosts.local.wd
						env: {
							CGO_ENABLED: "0"
							GOOS:        _os
							GOARCH:      _arch
						}
						cmd: "go"
						args: [
							"build",
							"-ldflags", strconv.Quote("-s -w -X github.com/octohelm/piper/internal/version.version=\(ver.version)"),
							"-o", "./.build/\(bin)_\(_os)_\(_arch)/\(bin)",
							"\(main)",
						]
					}

					_tar: {
						_contents: core.#Sub & {
							cwd:  _build.cwd
							path: "./.build/\(bin)_\(_os)_\(_arch)"
						}

						_tar: core.#Tar & {
							cwd:      hosts.local.wd
							path:     "./.build/\(bin)_\(_os)_\(_arch).tar.gz"
							contents: _contents.wd
						}
					}
				}
			}
		}
	}
}
