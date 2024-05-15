package golang

import (
	"path"
	"strings"
	"strconv"
	"regexp"

	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/exec"
	"piper.octohelm.tech/archive"
)

#Project: #ProjectBase & {
	cwd: wd.#WorkDir

	main:    _
	goos:    _
	goarch:  _
	env:     _
	ldflags: _
	bin:     _

	_info: #GoInfo & {
		gomod: wd: cwd
	}
	module: _info.output.module

	_out_dir: "./target"

	build: {
		for _os in goos for _arch in goarch {
			"\(_os)/\(_arch)": {
				_filename: "\(_out_dir)/\(bin)_\(_os)_\(_arch)/\(bin)"

				_run: exec.#Run & {
					"cwd": cwd
					"env": {
						env

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

				"file": file.#File & {
					wd: _run.cwd
					filename: {
						if _run.$ok {
							_filename
						}
					}
				}
			}
		}
	}

	"archive": {
		for _os in goos for _arch in goarch {
			"\(_os)/\(_arch)": {
				_out_file: build["\(_os)/\(_arch)"].file

				_dir: wd.#Sub & {
					cwd:    _out_file.wd
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

#ProjectBase: {
	main!:    string
	version!: string
	goos: [...string] | *["darwin", "linux"]
	goarch: [...string] | *["amd64", "arm64"]
	ldflags: [...string] | *["-s", "-w"]
	env: [Name=string]: string | client.#Secret
	bin: string | *path.Base(main)
	...
}

#GoInfo: {
	gomod: file.#File & {
		filename: "go.mod"
	}

	_read: file.#ReadAsString & {
		file: gomod
	}

	output: client.#Wait & {
		module:    regexp.FindSubmatch(#"module (.+)\n"#, _read.contents)[1]
		goversion: regexp.FindSubmatch(#"\ngo (.+)\n?"#, _read.contents)[1]
	}
}
