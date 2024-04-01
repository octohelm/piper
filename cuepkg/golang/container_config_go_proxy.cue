package golang

import (
	"piper.octohelm.tech/client"
	"piper.octohelm.tech/container"
)

#ConfigGoPrivate: {
	host!: string
	auth!: {
		username!: string
		secret!: client.#Secret
	}

	container.#Run &  {
		always: true
		env: {
			GIT_USERNAME:  auth.username
			GIT_TOKEN: auth.secret
		}
		run: """
		git config --global url.https://${GIT_USERNAME}:${GIT_TOKEN}@\(host)/.insteadOf https://\(host)/
		"""
	}
}


#ConfigGoPrivateForGitlabCI: {
	input: container.#Container

	_env: client.#Env & {
		CI_JOB_USER:  		string | *"gitlab-ci-token"
		CI_JOB_TOKEN!: 		client.#Secret
		CI_SERVER_HOST!:  string
	}

	_config: #ConfigGoPrivate & {
		"input": input
		"host":    _env.CI_SERVER_HOST
		"auth": {
			username: _env.CI_JOB_USER
			secret:   _env.CI_JOB_TOKEN
		}
	}

	output: _config.output
}