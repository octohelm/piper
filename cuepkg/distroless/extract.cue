package distroless

import (
	"strings"

	"piper.octohelm.tech/container"
	"github.com/octohelm/piper/cuepkg/debian"
)

#Extract: {
	platform!:     string
	debianversion: string | *debian.#DefaultVersion
	packages: [pkgName=string]: debian.#PackageOption

	include: [...string] | *[]
	exclude: [...string] | *[]

	_debian_image: debian.#DebianImage & {
		version: debianversion
	}

	_debian: debian.#Image & {
		source:     _debian_image.source
		"platform": platform
		"packages": packages
	}

	_pkg_path: {
		for pkgName, _ in packages {
			"\(pkgName)": {
				_run: container.#Run & {
					input: _debian.output
					run:   """
					dpkg -L \(pkgName) | xargs sh -c 'for f; do if [ -d "$f" ]; then echo "$f" >> /dirlist; else echo "$f" >> /filelist; fi done'
					"""
				}

				_dirlist: container.#ReadFile & {
					input: _run.output.rootfs
					path:  "/dirlist"
				}

				_filelist: container.#ReadFile & {
					input: _run.output.rootfs
					path:  "/filelist"
				}

				dirs:  strings.Split(strings.TrimSpace(_dirlist.contents), "\n")
				files: strings.Split(strings.TrimSpace(_filelist.contents), "\n")
			}
		}
	}

	_copy_files: container.#Sub & {
		input: _debian.output.rootfs
		"include": [
			for p in _pkg_path for f in p.files {
				strings.TrimPrefix(f, "/")
			},
			for f in include {
				f
			},
		]
		"exclude": exclude
	}

	_stretch: container.#Stretch & {
		"platform": platform
	}

	_mkdir: container.#Mkdir & {
		input: _stretch.output.rootfs
		path: [
			for p in _pkg_path for d in p.dirs {
				d
			},
		]
	}

	_merge: container.#Merge & {
		inputs: [
			_mkdir.output,
			_copy_files.output,
		]
	}

	output: _merge.output
}
