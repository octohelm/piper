package dagger

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestPiperRunnerHost(t *testing.T) {
	runnerHosts, err := ParsePiperRunnerHosts("docker-image://arm64builder@?platform=linux/arm64,docker-image://amd64builder@?platform=linux/amd64")
	testingx.Expect(t, err, testingx.Be[error](nil))
	t.Log(runnerHosts)
}
