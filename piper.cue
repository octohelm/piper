package main

import (
	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/container"
	"piper.octohelm.tech/kubepkg"

	"github.com/octohelm/piper/cuepkg/github"
	"github.com/octohelm/piper/cuepkg/golang"
	"github.com/octohelm/piper/cuepkg/debian"
	"github.com/octohelm/piper/cuepkg/distroless"
	"github.com/octohelm/piper/cuepkg/containerutil"

	"github.com/octohelm/piper/testdata/cuepkg/example"
)

hosts: local: wd.#Local & {
}

ver: client.#RevInfo & {
}

actions: go: golang.#Project & {
	cwd:     hosts.local.dir
	module:  _
	main:    "./cmd/piper"
	version: ver.version
	goos: [
		"darwin",
		"linux",
		"windows",
	]
	goarch: [
		"amd64",
		"arm64",
	]
	ldflags: [
		"-s", "-w",
		"-X", "\(module)/internal/version.version=\(version)",
	]
	env: GOEXPERIMENT: "rangefunc"
}

actions: release: {
	_env: client.#Env & {
		GH_PASSWORD!: client.#Secret
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

actions: ship: "distroless": distroless.#Ship & {
	name: "ghcr.io/octohelm/distroless"
	tag:  "static-debian-12"
}

actions: ship: piper: containerutil.#Ship & {
	name: "ghcr.io/octohelm/piper"
	tag:  "\(ver.version)"

	from: "docker.io/library/debian:bookworm-slim"

	steps: [
		debian.#InstallPackage & {
			packages: {
				git:  _
				make: _
				file: _
			}
		},
		{
			input: _

			_bin: container.#SourceFile & {
				file: actions.go.build[input.platform].file
			}

			_copy: container.#Copy & {
				"input":  input
				contents: _bin.output
				source:   "/"
				include: ["piper"]
				dest: "/usr/local/bin"
			}

			output: _copy.output
		},

		container.#Set & {
			config: {
				label: "org.opencontainers.image.source": "https://github.com/octohelm/piper"
				entrypoint: ["/bin/sh"]
			}
		},
	]
}

settings: {
	_env: client.#Env & {
		GH_USERNAME!: string
		GH_PASSWORD!: client.#Secret
	}

	registry: container.#Config & {
		auths: "ghcr.io": {
			username: _env.GH_USERNAME
			secret:   _env.GH_PASSWORD
		}
	}
}

actions: pkg: {
	manifests: kubepkg.#Manifests & {
		"kubepkg": example.#Example & {
		}
	}

	yaml: file.#WriteAsYAML & {
		data: manifests.manifests
		outFile: {
			wd:       hosts.local.dir
			filename: "target/\(manifests.kubepkg.metadata.name).kubepkg.yaml"
		}
	}

	ocitar: kubepkg.#OciTar & {
		"kubepkg": manifests.kubepkg
		"rename": "docker.io/x/{{ .name }}"
		"outFile": {
			wd:       hosts.local.dir
			filename: "target/\(manifests.kubepkg.metadata.name).kubepkg.tar"
		}
	}
}
