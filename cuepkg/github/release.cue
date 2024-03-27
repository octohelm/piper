package github

import (
	"path"
	"strings"
	"strconv"
	"encoding/json"

	"piper.octohelm.tech/http"
	"piper.octohelm.tech/file"
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/flow"
)

#GithubAPI: {
	core:    "https://api.github.com"
	uploads: "https://uploads.github.com"
}

#Release: {
	token: client.#Secret

	owner: string
	repo:  string

	name:       string | *"latest"
	notes:      string | *""
	prerelease: bool | *true
	draft:      bool | *false

	assets: [...file.#File]

	_client: #Client & {"token": token}

	_get_or_create_release: {
		_ret: flow.#Some & {
			steps: [
				// get release by tag
				_client.#Do & {
					method: "GET"
					url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/tags/\(name)"
				},
				// or
				// create an new release
				_client.#Do & {
					"method": "POST"
					"url":    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases"
					"header": {
						"Content-Type": "application/json"
					}
					"body": json.Marshal({
						"tag_name":   name
						"name":       name
						"body":       notes
						"prerelease": prerelease
						"draft":      draft
					})
				},
			]
		}

		id: [
			if _ret.ok {
				for step in _ret.condition if step.ok {
					step.response.data.id
				}
			},
		][0]
	}

	_list_assets: {
		_req: _client.#Do & {
			method: "GET"
			url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/\(_get_or_create_release.id)/assets"
		}

		assets: {
			for asset in _req.response.data {
				"\(asset.name)": strconv.FormatFloat(asset.id, strings.ByteAt("f", 0), 0, 64)
			}
		}
	}

	_upload_assets: flow.#Every & {
		steps: [
			for f in assets {
				let assetName = path.Base(f.filename)

				flow.#Every & {
					steps: [
						// if asset name exists, delete first
						if _list_assets.assets[assetName] != _|_ {
							_client.#Do & {
								"method": "DELETE"
								"url":    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/assets/\(_list_assets.assets[assetName])"
							}
						},
						// then upload
						_client.#Do & {
							"method": "POST"
							"url":    "\(#GithubAPI.uploads)/repos/\(owner)/\(repo)/releases/\(_get_or_create_release.id)/assets"
							"header": {
								"Content-Type": "application/octet-stream"
							}
							"query": {
								"name": "\(assetName)"
							}
							"body": f
						},
					]
				}
			},
		]
	}

	ok: _upload_assets.ok
}

#Client: {
	token: client.#Secret

	_token: client.#ReadSecret & {
		secret: token
	}

	#Do: {
		method: string
		url:    string
		body?:  file.#StringOrFile
		header: [Name=string]: string | [...string]
		query: [Name=string]: string | [...string]

		_default_header: {
			"Accept":               "application/vnd.github+json"
			"Authorization":        "Bearer \(_token.value)"
			"X-GitHub-Api-Version": "2022-11-28"
		}

		_req: http.#Do & {
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
			if body != _|_ {
				"body": body
			}
		}

		ok:       _req.ok
		response: _req.response
	}
}
