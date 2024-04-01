package bun

import "github.com/octohelm/piper/cuepkg/debian"

#Image: {
	bunversion: string | *"1"

	debian.#Image & {
		source: "docker.io/oven/bun:\(bunversion)"
	}
}
