package example

import (
	kubepkg "github.com/octohelm/kubepkgspec/cuepkg/kubepkg"
)

#Example: kubepkg.#KubePkg & {
	metadata: {
		name:      "demo"
		namespace: "default"
	}
	spec: {
		version: "0.0.2"
		config: X: "x"
		deploy: {
			kind: "Deployment"
			spec: replicas: 1
		}
		containers: web: {
			image: {
				name:       "docker.io/library/nginx"
				tag:        "1.25.0-alpine"
				pullPolicy: "IfNotPresent"
				platforms: [
					"linux/amd64",
					"linux/arm64",
				]
			}
			ports: http: 80
		}
		services: "#": {
			ports: http: 80
			paths: http: "/"
		}
		volumes: html: {
			mountPath: "/usr/share/nginx/html"
			type:      "ConfigMap"
			spec: data: "index.html": "<div>hello</div>"
		}
	}
}
