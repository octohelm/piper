package sshutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/k0sproject/rig"
	"github.com/kevinburke/ssh_config"

	"github.com/octohelm/x/ptr"
)

func Load(configPath string, hostKey string) (*rig.SSH, error) {
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".ssh", "config")
	}

	if strings.HasPrefix(configPath, "~/") {
		configPath = filepath.Join(os.Getenv("HOME"), configPath[2:])
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	for _, host := range cfg.Hosts {
		if len(host.Patterns) > 0 && host.Patterns[0].String() != "*" {
			if host.Matches(hostKey) {
				ssh := &rig.SSH{}

				for _, node := range host.Nodes {
					switch x := node.(type) {
					case *ssh_config.KV:
						switch strings.ToLower(x.Key) {
						case "hostname":
							ssh.Address = x.Value
						case "port":
							ssh.Port, _ = strconv.Atoi(x.Value)
						case "user":
							ssh.User = x.Value
						case "identityfile":
							ssh.KeyPath = ptr.Ptr(mayUnquote(x.Value))
						}
					}
				}
				return ssh, nil
			}
		}
	}

	for _, host := range cfg.Hosts {
		if host.Matches("*") {
			for _, node := range host.Nodes {
				switch x := node.(type) {
				case *ssh_config.Include:
					return Load(x.String()[len("Include "):], hostKey)
				}
			}
		}
	}

	return nil, fmt.Errorf("not found %s", hostKey)
}

func mayUnquote(s string) string {
	if s != "" && s[0] == '"' {
		v, _ := strconv.Unquote(s)
		return v
	}
	return s
}
