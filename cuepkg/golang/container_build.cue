package golang

import (
	"strings"
	"path"

	"piper.octohelm.tech/client"
	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/container"

	"github.com/octohelm/piper/cuepkg/debian"
)

#GolangImage: {
	version!: string
	name:     string | *"docker.io/library/golang"
	source:   "docker.io/library/golang:\(version)-\(debian.#DefaultVersion)"
}

#Image: {
	from!: string

	debian.#Image & {
		"source": "\(from)"
	}
}

#ContainerBuild: #ProjectBase & {
	source: container.#Source

	_info: #GoInfo & {
		gomod: wd: source.cwd
	}
	module: _info.output.module

	_goenv: client.#Env & {
		GOPROXY:   string | *""
		GOPRIVATE: string | *""
		GOSUMDB:   string | *""
	}

	main:    _
	goos:    _
	goarch:  _
	env:     _
	ldflags: _
	bin:     _

	mounts: [Name=string]: container.#Mount

	env: {
		GOPROXY:   _goenv.GOPROXY
		GOPRIVATE: _goenv.GOPRIVATE
		GOSUMDB:   _goenv.GOSUMDB
	}

	_default_image: #GolangImage & {
		version: "\(_info.output.goversion)"
	}

	// build image
	image: #Image & {
		from: _ | *_default_image.source
	}

	build: {
		workdir: string | *"/go/src"
		prepare: client.#StringOrSlice | *"go mod download -x"

		_cache: {
			go_mod_cache:   "/go/pkg/mod"
			go_build_cache: "/root/.cache/go-build"
		}

		_cached_mounts: {
			for _n, _p in _cache {
				"\(_p)": container.#Mount & {
					dest: _p
					contents: container.#CacheDir & {
						id: "\(_n)"
					}
				}
			}
		}

		_load_source: container.#Build & {
			steps: [
				container.#Copy & {
					"input":    image.output
					"contents": source.output
					"dest":     "\(workdir)"
				},
				if prepare != _|_ {
					container.#Run & {
						"workdir": "\(workdir)"
						"mounts":  _cached_mounts
						"run":     prepare
						"env":     env
					}
				},
			]
		}

		for _os in goos for _arch in goarch {
			"\(_os)/\(_arch)": {
				_outDir: "./target"

				_build: container.#Run & {
					"input":   _load_source.output
					"workdir": "\(workdir)"
					"run":     "go build -ldflags=\"\(strings.Join(ldflags, " "))\" -o \(_outDir)/\(bin) \(main)"
					"env": {
						env
						CGO_ENABLED: "0"
						GOOS:        "\(_os)"
						GOARCH:      "\(_arch)"
					}
					"mounts": {
						mounts
						_cached_mounts
					}
				}

				_dist: container.#Sub & {
					input:  _build.output.rootfs
					source: "\(path.Join([workdir, _outDir]))"
					dest:   "/"
				}

				output: _dist.output
			}
		}
	}

	dump: {
		outDir: string | *"./target"

		for _os in goos for _arch in goarch {
			"\(_os)/\(_arch)": {
				_outDir: wd.#Sub & {
					"cwd": source.cwd
					"path": path.Join([outDir, "\(bin)_\(_os)_\(_arch)"])
				}

				_dump: container.#Dump & {
					input: build["\(_os)/\(_arch)"].output
					with: {
						empty: true
					}
					outDir: _outDir.dir
				}

				"file": file.#File & {
					wd:       _dump.dir
					filename: bin
				}
			}
		}
	}
}
