package wd

func GoOS(id string) (string, bool) {
	if id == "darwin" || id == "windows" {
		return id, true
	}
	return "linux", true
}

func GoArch(os string, unameM string) (string, bool) {
	switch os {
	case "windows":
		v, ok := windowsArches[unameM]
		return v, ok
	case "linux":
		v, ok := linuxArches[unameM]
		return v, ok
	case "darwin":
		v, ok := darwinArches[unameM]
		return v, ok
	}
	return "", false
}

var windowsArches = map[string]string{
	"x86_64":  "amd64",
	"aarch64": "arm64",
}

var linuxArches = map[string]string{
	"x86_64":  "amd64",
	"aarch64": "arm64",
}

var darwinArches = map[string]string{
	"x86_64": "amd64",
	"arm64":  "arm64",
}
