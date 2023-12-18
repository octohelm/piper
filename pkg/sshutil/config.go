package sshutil

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/k0sproject/rig"
	"github.com/kevinburke/ssh_config"
	"github.com/pkg/errors"
)

func Load(configPath string, alias string) (*rig.SSH, error) {
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
			if host.Matches(alias) {
				ssh := &rig.SSH{}

				for _, node := range host.Nodes {
					switch x := node.(type) {
					case *ssh_config.KV:
						switch x.Key {
						case "Hostname":
							ssh.Address = x.Value
						case "Port":
							ssh.Port, _ = strconv.Atoi(x.Value)
						case "User":
							ssh.User = x.Value
						case "IdentityFile":
							v, _ := strconv.Unquote(x.Value)
							ssh.KeyPath = &v
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
					return Load(x.String()[len("Include "):], alias)
				}
			}
		}
	}

	return nil, errors.Errorf("not found %s", alias)
}
