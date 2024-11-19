package github

import (
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/http"
	"piper.octohelm.tech/client"
)

#GithubAPI: {
	core:    "https://api.github.com"
	uploads: "https://uploads.github.com"
}

#Client: {
	token: client.#Secret

	_token: client.#ReadSecret & {
		secret: token
	}

	#Do: {
		$dep?: _

		method: string
		url:    string
		body?:  file.#StringOrFile
		header: [Name=string]: string | [...string]
		query: [Name=string]: string | [...string]

		_default_header: {
			Accept:                 "application/vnd.github+json"
			Authorization:          "Bearer \(_token.value)"
			"X-GitHub-Api-Version": "2022-11-28"
		}

		_req: http.#Do & {
			// to avoid lost deps
			if $dep != _|_ {
				"$dep": $dep
			}

			"method": method
			"url":    url
			"header": {
				for k, vv in header {
					"\(k)": vv
				}
				for k, vv in _default_header if header[k] == _|_ {
					"\(k)": vv
				}
			}
			"query": query

			// to avoid lost deps
			if body != _|_ {
				"body": body
			}
		}

		$ok: _req.$ok

		response: http.#Response & {
			if _req.response.status != _|_ {
				status: _req.response.status
			}
			if _req.response.header != _|_ {
				header: _req.response.header
			}
			if _req.response.header != _|_ {
				data: _req.response.data
			}
		}
	}
}
