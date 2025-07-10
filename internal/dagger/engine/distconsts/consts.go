package distconsts

const (
	EngineContainerName = "dagger-engine.dev"

	DefaultEngineSockAddr = "unix:///run/dagger/engine.sock"
)

const (
	RuncPath       = "/usr/local/bin/runc"
	DaggerInitPath = "/usr/local/bin/dagger-init"

	EngineDefaultStateDir = "/var/lib/dagger"

	EngineContainerBuiltinContentDir   = "/usr/local/share/dagger/content"
	GoSDKManifestDigestEnvName         = "DAGGER_GO_SDK_MANIFEST_DIGEST"
	PythonSDKManifestDigestEnvName     = "DAGGER_PYTHON_SDK_MANIFEST_DIGEST"
	TypescriptSDKManifestDigestEnvName = "DAGGER_TYPESCRIPT_SDK_MANIFEST_DIGEST"
)

const (
	AlpineVersion = "3.22.0"
	AlpineImage   = "alpine:" + AlpineVersion

	GolangVersion = "1.24.4"
	GolangImage   = "golang:" + GolangVersion + "-alpine"

	BusyboxVersion = "1.37.0"
	BusyboxImage   = "busybox:" + BusyboxVersion
)

const (
	OCIVersionAnnotation = "org.opencontainers.image.version"
)
