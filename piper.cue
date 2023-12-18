package main

import (
	"piper.octohelm.tech/wd"
	"piper.octohelm.tech/client"

	"github.com/octohelm/piper/cuepkg/github"
	"github.com/octohelm/piper/cuepkg/golang"
)

hosts: {
	local: wd.#Local & {
	}
}

ver: client.#RevInfo & {
}

actions: go: golang.#Project & {
	cwd:  hosts.local.wd
	main: "./cmd/piper"
	os: [
		"darwin",
		"linux",
		"windows",
	]
	arch: [
		"amd64",
		"arm64",
	]
	ldflags: [
		"-s", "-w",
		"-X", "github.com/octohelm/piper/internal/version.version=\(ver.version)",
	]
}

actions: release: {
	_env: client.#Env & {
		GH_PASSWORD: client.#Secret
	}

	github.#Release & {
		owner:      "octohelm"
		repo:       "piper"
		token:      _env.GH_PASSWORD
		prerelease: true
		assets: [
			for f in actions.go.archive {
				f.result.file
			},
		]
	}
}
