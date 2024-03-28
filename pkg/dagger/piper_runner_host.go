package dagger

import (
	"fmt"
	pkgwd "github.com/octohelm/piper/pkg/wd"
	"net/url"
	"strings"
)

func ParsePiperRunnerHosts(str string) ([]PiperRunnerHost, error) {
	settings := strings.Split(str, ",")

	runnerHosts := make([]PiperRunnerHost, 0, len(settings))

	for i, e := range settings {
		u, err := url.Parse(e)
		if err != nil {
			return nil, err
		}

		runnerHost := PiperRunnerHost{}

		if u.Host == "default" || u.Host == "" {
			runnerHost.RunnerHost = DefaultRunnerHost
		} else {
			runnerHost.RunnerHost = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}

		if u.User != nil {
			runnerHost.Name = u.User.Username()
		} else {
			runnerHost.Name = fmt.Sprintf("piper-runner-%d", i)
		}

		if platforms, ok := u.Query()["platform"]; ok {
			for _, platform := range platforms {
				p, err := pkgwd.ParsePlatform(platform)
				if err != nil {
					return nil, err
				}
				runnerHost.Platforms = append(runnerHost.Platforms, *p)
			}
		}

		runnerHosts = append(runnerHosts, runnerHost)
	}

	return runnerHosts, nil
}

type PiperRunnerHost struct {
	Name       string
	RunnerHost string
	Platforms  []pkgwd.Platform
}
