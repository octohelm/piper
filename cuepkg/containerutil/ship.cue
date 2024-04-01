package containerutil

import (
	"piper.octohelm.tech/container"
)

#Ship: {
	name!: string
	tag!:  string
	from!: string
	platforms: [...string] | *["linux/amd64", "linux/arm64"]
	steps: [...container.#Step]

	build: {
		for _platform in platforms {
			"\(_platform)": container.#Build & {
				"steps": [
					container.#Pull & {
						"source":   from
						"platform": _platform
					},
					for step in steps {
						step
					},
				]
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
