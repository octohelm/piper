package debian

import (
	"piper.octohelm.tech/container"
)

#DefaultVersion: "bookworm"

#DebianImage: #ImageSource & {
	name:    string | *"docker.io/library/debian"
	version: string | *"\(#DefaultVersion)" // debian 12
	source:  "\(name):\(version)-slim"
}

#ImageSource: {
	source: string
	...
}

#Image: {
	let _defaultSource = #DebianImage

	source:    string | *_defaultSource.source
	platform?: string
	packages: [pkgName=string]: #PackageOption
	steps: [...container.#Step]

	_build: container.#Build & {
		"steps": [
			container.#Pull & {
				"source": source
				if platform != _|_ {
					"platform": platform
				}
			},
			#InstallPackage & {
				input:      _
				"packages": packages
			},
			for step in steps {
				step
			},
		]
	}

	output: _build.output
}
