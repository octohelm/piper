package golang

import (
	"path"
	"strings"
	"strconv"

	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/exec"
	"piper.octohelm.tech/archive"
)

#Project: {
	cwd:  wd.#WorkDir
	main: string
	os: [...string] | *["darwin", "linux"]
	arch: [...string] | *["amd64", "arm64"]
	ldflags: [...string] | *["-s", "-w"]

	bin: string | *path.Base(main)

	_buildDir: "./.build"

	build: {
		for _os in os for _arch in arch {
			"\(_os)/\(_arch)": {
				_filename: "\(_buildDir)/\(bin)_\(_os)_\(_arch)/\(bin)"

				_run: exec.#Run & {
					"cwd": cwd
					env: {
						CGO_ENABLED: "0"
						GOOS:        _os
						GOARCH:      _arch
					}
					cmd: [
						"go", "build",
						"-ldflags", strconv.Quote(strings.Join(ldflags, " ")),
						"-o", _filename,
						"\(main)",
					]
				}

				result: {
					ok: _run.result.ok

					if _run.result.ok {
						"file": file.#File & {
							wd:      _run.cwd
							filename: _filename
						}
					}
				}
			}
		}
	}

	"archive": {
		for _os in os for _arch in arch {
			"\(_os)/\(_arch)": {
				_built_file: build["\(_os)/\(_arch)"].result.file

				_dir: wd.#Sub & {
					"cwd":  _built_file.wd
					"path": path.Dir(_built_file.filename)
				}

				_tar: archive.#Tar & {
					cwd:      _built_file.wd
					dir:      _dir.wd
					outFile: "\(_buildDir)/\(bin)_\(_os)_\(_arch).tar.gz"
				}

				result: _tar.result
			}
		}
	}
}
