package debian

import (
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/container"
)

#PackageOption: {
	platform?: string
	version:   string | *""
}

#InstallPackage: {
	input: container.#Container
	packages: [pkgName=string]: #PackageOption

	_install: container.#Build & {
		steps: [
			if len(packages) == 0 {
				output: input
			},

			if len(packages) > 0 {
				_client_env: client.#Env & {
					LINUX_MIRROR: _ | *""
				}

				_mirror_configrated: container.#Run & {
					"input": input
					"env": {
						LINUX_MIRROR: _client_env.LINUX_MIRROR
					}
					"run": """
						if [ "${LINUX_MIRROR}" != "" ]; then
								if [ -f "/etc/apt/sources.list" ]; then
										sed -i "s@http://deb.debian.org@${LINUX_MIRROR}@g" /etc/apt/sources.list
										sed -i "s@http://security.debian.org@${LINUX_MIRROR}@g" /etc/apt/sources.list
								fi
								if [ -f "/etc/apt/sources.list.d/debian.sources" ]; then
										sed -i "s@http://deb.debian.org@${LINUX_MIRROR}@g" /etc/apt/sources.list.d/debian.sources
										sed -i "s@http://security.debian.org@${LINUX_MIRROR}@g" /etc/apt/sources.list.d/debian.sources
								fi
						fi
						"""
				}

				_dirs: {
					"varlog":    "/var/log"
					"apt_cache": "/var/apt/cache"
				}

				_run: container.#Run & {
					"input": _mirror_configrated.output
					"mounts": {
						for id, dir in _dirs {
							"\(id)": container.#Mount & {
								dest: "\(dir)"
								contents: container.#CacheDir & {
									"id": "\(input.platform)/\(id)"
								}
							}
						}
					}
					"run": [
						"apt-get update -y",
						for _pkgName, _opt in packages {
							[
								// only install platform matched
								if _opt.platform != _|_ if _opt.platform == input.platform {
									"apt-get install -y -f \(_pkgName)\(_opt.version)"
								},
								"apt-get install -y -f \(_pkgName)\(_opt.version)",
							][0]
						},
					]
				}

				output: _run.output
			},
		]
	}

	output: _install.output
}
