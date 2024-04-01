package golang

import (
	"path"
	"piper.octohelm.tech/container"
)

#ContainerBuild: {
	source: container.#Source

	main!: string
	os: [...string] | *["darwin", "linux"]
	arch: [...string] | *["amd64", "arm64"]
	ldflags: [...string] | *["-s", "-w"]
	bin: string | *path.Base(main)
}
