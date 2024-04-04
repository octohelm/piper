package distroless

import (
	"piper.octohelm.tech/container"
	"piper.octohelm.tech/wd"
)

#Ship: {
	name!: string
	tag:   string | *"static-debian-12"

	platforms: ["linux/amd64", "linux/arm64"]

	build: {
		for _platform in platforms {
			"\(_platform)": #Static & {
				platform: _platform
			}
		}
	}

	_tmp: wd.#Temp & {
		id: "distroless"
	}

	export: {
		for _platform in platforms {
			"\(_platform)": container.#Dump & {
				input:  build["\(_platform)"].output.rootfs
				outDir: _tmp.dir
			}
		}
	}

	push: container.#Push & {
		dest: "\(name):\(tag)"
		images: {
			for _platform in platforms {
				"\(_platform)": build["\(_platform)"].output
			}
		}
	}
}
