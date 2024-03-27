package debian

import (
	"piper.octohelm.tech/container"
)

#ImageBase: {
	packages: [pkgName=string]: #PackageOption
	steps: [...container.#Step]
	...
}

#Image: #ImageBase & {
	debianversion: string | *"bookworm" // debian 12
	source:        string | *"docker.io/library/debian:\(debianversion)-slim"
	platform?:     string

	packages: _
	steps:    _

	_build: container.#Build & {
		"steps": [
			container.#Pull & {
				"source": source
				if platform != _|_ {
					"platform": platform
				}
			},
			#InstallPackage & {
				"input":    _
				"packages": packages
			},
			for step in steps {
				step
			},
		]
	}

	output: _build.output
}
