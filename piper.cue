package main

import (
	"path"

	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/container"

	"github.com/octohelm/piper/cuepkg/github"
	"github.com/octohelm/piper/cuepkg/golang"
	"github.com/octohelm/piper/cuepkg/debian"
)

hosts: {
	local: wd.#Local & {
	}
}

ver: client.#RevInfo & {
}

actions: go: golang.#Project & {
	cwd:  hosts.local.dir
	main: "./cmd/piper"
	os: [
		"darwin",
		"linux",
		"windows",
	]
	arch: [
		"amd64",
		"arm64",
	]
	ldflags: [
		"-s", "-w",
		"-X", "github.com/octohelm/piper/internal/version.version=\(ver.version)",
	]
}

actions: release: {
	_env: client.#Env & {
		GH_PASSWORD: client.#Secret
	}

	github.#Release & {
		owner:      "octohelm"
		repo:       "piper"
		token:      _env.GH_PASSWORD
		prerelease: true
		assets: [
			for f in actions.go.archive {
				f.file
			},
		]
	}
}

settings: {
	_env: client.#Env & {
		GH_USERNAME: string | *""
		GH_PASSWORD: client.#Secret
	}

	registry: container.#Config & {
		auths: "ghcr.io": {
			username: _env.GH_USERNAME
			secret:   _env.GH_PASSWORD
		}
	}
}

actions: ship: {
	arch: [
		"amd64",
		"arm64",
	]

	build: {
		for a in arch {
			"linux/\(a)": {
				_built_file: actions.go.build["linux/\(a)"].file

				_bin: container.#Source & {
					"cwd":  _built_file.wd
					"path": path.Dir(_built_file.filename)
					"include": [
						path.Base(_built_file.filename),
					]
				}

				debian.#Image & {
					platform: "linux/\(a)"
					packages: {
						"git":  _
						"wget": _
						"curl": _
						"make": _
					}
					steps: [
						container.#Copy & {
							contents: _bin.output
							source:   "/"
							dest:     "/"
						},
						container.#Set & {
							config: {
								label: "org.opencontainers.image.source": "https://github.com/octohelm/piper"
								entrypoint: ["/bin/sh"]
							}
						},
					]
				}
			}
		}
	}

	publish: container.#Push & {
		dest: "ghcr.io/octohelm/piper:\(ver.version)"
		images: {
			for a in arch {
				"linux/\(a)": build["linux/\(a)"].output
			}
		}
	}
}
