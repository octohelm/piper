package container

import (
	"dagger.io/dagger"

	piperdagger "github.com/octohelm/piper/pkg/dagger"
)

type ImageConfig piperdagger.ImageConfig

func (p ImageConfig) ApplyTo(c *dagger.Container) *dagger.Container {
	for k := range p.Env {
		c = c.WithEnvVariable(k, p.Env[k])
	}
	for k := range p.Labels {
		c = c.WithLabel(k, p.Labels[k])
	}
	if vv := p.User; vv != "" {
		c = c.WithUser(vv)
	}
	if vv := p.WorkingDir; vv != "" {
		c = c.WithWorkdir(vv)
	}
	if vv := p.Entrypoint; len(vv) != 0 {
		c = c.WithEntrypoint(vv)
	}
	if vv := p.Cmd; len(vv) != 0 {
		c = c.WithDefaultArgs(vv)
	}
	return c
}

func (p ImageConfig) Merge(config ImageConfig) ImageConfig {
	final := ImageConfig{}

	final.WorkingDir = p.WorkingDir
	if config.WorkingDir != "" {
		final.WorkingDir = config.WorkingDir
	}

	final.User = p.User
	if config.WorkingDir != "" {
		final.User = config.User
	}

	final.Entrypoint = p.Entrypoint
	if len(config.Entrypoint) != 0 {
		final.Entrypoint = config.Entrypoint
	}

	final.Cmd = p.Cmd
	if len(config.Cmd) != 0 {
		final.Cmd = config.Cmd
	}

	final.Labels = MergeMap(p.Labels, config.Labels)
	final.Env = MergeMap(p.Env, config.Env)

	return final
}

func MergeMap[K comparable, V any](a map[K]V, b map[K]V) map[K]V {
	m := make(map[K]V)

	for k := range a {
		m[k] = a[k]
	}

	for k := range b {
		m[k] = b[k]
	}

	return m
}
