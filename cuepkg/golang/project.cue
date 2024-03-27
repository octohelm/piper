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
	cwd: wd.#WorkDir

	main: string
	os: [...string] | *["darwin", "linux"]
	arch: [...string] | *["amd64", "arm64"]
	ldflags: [...string] | *["-s", "-w"]

	bin: string | *path.Base(main)

	_out_dir: "./target"

	build: {
		for _os in os for _arch in arch {
			"\(_os)/\(_arch)": {
				_filename: "\(_out_dir)/\(bin)_\(_os)_\(_arch)/\(bin)"

				_run: exec.#Run & {
					"cwd": cwd
					"env": {
						CGO_ENABLED: "0"
						GOOS:        _os
						GOARCH:      _arch
					}
					"cmd": [
						"go", "build",
						"-ldflags", strconv.Quote(strings.Join(ldflags, " ")),
						"-o", _filename,
						"\(main)",
					]
				}

				"file": {
					if _run.ok {
						file.#File & {
							wd:       _run.cwd
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
				_out_file: build["\(_os)/\(_arch)"].file

				_dir: wd.#Sub & {
					"cwd":  _out_file.wd
					"path": path.Dir(_out_file.filename)
				}

				_tar: archive.#Tar & {
					srcDir: _dir.dir
					outFile: {
						wd:       cwd
						filename: "\(_out_dir)/\(bin)_\(_os)_\(_arch).tar.gz"
					}
				}

				file: _tar.file
			}
		}
	}
}
