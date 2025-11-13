package golang

import "piper.octohelm.tech/container"

#InstallZig: X = {
	input: container.#Container

	_install: container.#Run & {
		input: X.input
		env: {
			ZIG_VERSION: "0.15.2"
		}
		run: """
			set -e
			
			TMP=$(mktemp -d)
			
			URL="https://ziglang.org/download/${ZIG_VERSION}/zig-$(uname -m)-linux-${ZIG_VERSION}.tar.xz"
			
			cd "$TMP"
			curl -LO "$URL"
			tar -xJf *.tar.xz
			cd zig-*
			cp zig /usr/local/bin/
			rm -rf /usr/local/lib/zig
			cp -r lib /usr/local/lib/zig
			cd /
			rm -rf "$TMP"
			zig version
			"""
	}

	output: _install.output
}
