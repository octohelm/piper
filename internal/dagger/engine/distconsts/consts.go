package distconsts

const (
	RuncPath     = "/usr/local/bin/runc"
	DumbInitPath = "/usr/local/bin/dumb-init"

	EngineDefaultStateDir = "/var/lib/dagger"

	EngineContainerBuiltinContentDir   = "/usr/local/share/dagger/content"
	GoSDKManifestDigestEnvName         = "DAGGER_GO_SDK_MANIFEST_DIGEST"
	PythonSDKManifestDigestEnvName     = "DAGGER_PYTHON_SDK_MANIFEST_DIGEST"
	TypescriptSDKManifestDigestEnvName = "DAGGER_TYPESCRIPT_SDK_MANIFEST_DIGEST"
)

const (
	AlpineVersion = "3.20.0"
	AlpineImage   = "alpine:" + AlpineVersion

	GolangVersion = "1.22.4"
	GolangImage   = "golang:" + GolangVersion + "-alpine"
)
