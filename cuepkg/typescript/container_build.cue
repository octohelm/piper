package typescript

import (
	"path"
	"strings"

	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/container"

	"github.com/octohelm/piper/cuepkg/debian"
)

#BunImageResolver: {
	packagejson: file.#File & {
		filename: "package.json"
	}

	_read: file.#ReadFromJSON & {
		file: packagejson
	}

	_output: client.#Wait & {
		image: #BunImage & {
			version: "\(strings.Split(_read.data.packageManager, "@")[1])"
		}
	}

	image: _output.image
}

#BunImage: {
	name:    string | *"docker.io/oven/bun"
	version: string | *"1"

	source: "\(name):\(version)"
}

#NodeImage: {
	nodeversion: string | *"21"

	name:    string | *"library/node"
	version: string | *"21"

	source: "\(name):\(version)"
}

#Image: {
	let _default_image = #BunImage

	from: string | *(_default_image.source)

	debian.#Image & {
		source: "\(from)"
	}
}

#ContainerBuild: {
	source: container.#Source

	image: #Image

	build: {
		workdir: "/app"
		run!:    client.#StringOrSlice

		env: [Key=string]:     string | container.#Secret
		mounts: [Name=string]: container.#Mount

		outDir: string | *"."

		_load_source: container.#Copy & {
			input:    image.output
			contents: source.output
			dest:     "\(workdir)"
		}

		_run: container.#Run & {
			input: _load_source.output
			"mounts": {
				mounts

				bun_install_cache: container.#Mount & {
					type: "cache"
					// https://bun.sh/docs/install/cache
					dest: "/root/.bun/install/cache"
					contents: container.#CacheDir & {
						id: "bun_install_cache"
					}
				}
			}
			"env":     env
			"workdir": "\(workdir)"
			"run":     run
		}

		_dist: container.#Sub & {
			input:  _run.output.rootfs
			source: "\(path.Join([workdir, outDir]))"
			dest:   "/"
		}

		output: _dist.output
	}

	dump: {
		outDir: string | *"target/bun"

		_outDir: wd.#Sub & {
			cwd:  source.cwd
			path: outDir
		}

		_dump: container.#Dump & {
			input: build.output
			with: empty: true
			outDir: _outDir.dir
		}

		dir: _dump.dir
	}
}
