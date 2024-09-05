package github

import (
	"path"
	"strings"
	"strconv"
	"encoding/json"

	"piper.octohelm.tech/file"
	"piper.octohelm.tech/client"
)

#Release: {
	token: client.#Secret

	owner: string
	repo:  string

	name:       string | *"latest"
	notes:      string | *""
	prerelease: bool | *true
	draft:      bool | *false

	assets: [...file.#File]

	_client: #Client & {
		"token": token
	}

	upload_assets: client.#Group & {
		_get_or_create_release: {
			_get_release: _client.#Do & {
				$dep: assets

				method: "GET"
				url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/tags/\(name)"
			}

			_create_release: _client.#Do & {
				$dep: client.#Skip & {when: _get_release.$ok}
				method: "POST"
				url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases"
				header: "Content-Type": "application/json"
				body: json.Marshal({
					tag_name:     name
					"name":       name
					body:         notes
					"prerelease": prerelease
					"draft":      draft
				})
			}

			_steps: [
				_get_release,
				_create_release,
			]

			_select: client.#Wait & {
				release_id: [
					for _step in _steps if _step.$ok != _|_ && _step.$ok {
						_step.response.data.id
					},
				][0]
			}

			release_id: "\(_select.release_id)"
		}

		_list_assets: {
			_req: _client.#Do & {
				method: "GET"
				url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/\(_get_or_create_release.release_id)/assets"
			}

			asset_ids: client.#Wait & {
				for _asset in _req.response.data {
					"\(_asset.name)": strconv.FormatFloat(_asset.id, strings.ByteAt("f", 0), 0, 64)
				}
			}
		}

		upload: {
			for f in assets {
				let assetName = path.Base(f.filename)

				"\(assetName)": {
					_delete_old_if_exists: _client.#Do & {
						$dep: client.#Skip & {
							when: (_list_assets.asset_ids[assetName] == _|_)
						}
						method: "DELETE"
						url:    "\(#GithubAPI.core)/repos/\(owner)/\(repo)/releases/assets/\(_list_assets.asset_ids[assetName])"
					}

					_upload: _client.#Do & {
						$dep:   _delete_old_if_exists.$ok
						method: "POST"
						url:    "\(#GithubAPI.uploads)/repos/\(owner)/\(repo)/releases/\(_get_or_create_release.release_id)/assets"
						header: "Content-Type": "application/octet-stream"
						query: name:            "\(assetName)"
						body: f
					}

					$ok: _upload.$ok
				}
			}
		}
	}
}
