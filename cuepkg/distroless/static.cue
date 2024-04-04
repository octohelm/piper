package distroless

import (
	"piper.octohelm.tech/container"
)

#Static: {
	platform!: string

	_base_files: #BaseFiles & {
		"platform": platform
	}

	_netbase: #Netbase & {
		"platform": platform
	}

	_tzdata: #Tzdata & {
		"platform": platform
	}

	_ca_certificates: #CaCertificates & {
		"platform": platform
	}

	_build: container.#Build & {
		steps: [
			container.#Stretch & {
				"platform": platform
			},
			container.#Copy & {
				contents: _base_files.output
			},
			container.#Copy & {
				contents: _netbase.output
			},
			container.#Copy & {
				contents: _tzdata.output
			},
			container.#Copy & {
				contents: _ca_certificates.output
			},
			container.#RootfsDo & {
				steps: [
					#EtcGroup,
					#EtcPasswd,
					#EtcNsswitch,
				]
			},
			container.#Set & {
				config: {
					env: {
						"PATH":          "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
						"SSL_CERT_FILE": "/etc/ssl/certs/ca-certificates.crt"
					}
				}
			},
		]
	}

	output: _build.output
}
