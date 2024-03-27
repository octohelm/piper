package dagger

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

type ImageConfig struct {
	WorkingDir string            `json:"workdir" default:""`
	Env        map[string]string `json:"env,omitempty"`
	Labels     map[string]string `json:"label,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Cmd        []string          `json:"cmd,omitempty"`
	User       string            `json:"user" default:""`
}

func ResolveImageConfig(ctx context.Context, c *Client, id dagger.ContainerID) (ImageConfig, error) {
	ret := struct {
		Container struct {
			ID dagger.ContainerID `json:"id,omitempty"`

			Platform string `json:"platform,omitempty"`

			Entrypoint  []string `json:"entrypoint,omitempty"`
			DefaultArgs []string `json:"defaultArgs,omitempty"`
			Workdir     string   `json:"workdir,omitempty"`

			User         string        `json:"user,omitempty"`
			EnvVariables []EnvVariable `json:"envVariables,omitempty"`
			Labels       []Label       `json:"labels,omitempty"`
		} `json:"container"`
	}{}

	err := query(ctx, c, &ret, fmt.Sprintf(`
query { 
    container(id: %q) {
		id
		entrypoint
		defaultArgs
		workdir
		user
		labels {
			value
			name
		}
		envVariables {
			value
			name
		}
    }
}
`, id))
	if err != nil {
		return ImageConfig{}, err
	}

	p := ImageConfig{}

	p.WorkingDir = ret.Container.Workdir
	p.Entrypoint = ret.Container.Entrypoint
	p.Cmd = ret.Container.DefaultArgs
	p.User = ret.Container.User

	p.Env = map[string]string{}
	for _, e := range ret.Container.EnvVariables {
		p.Env[e.Name] = e.Value
	}

	p.Labels = map[string]string{}
	for _, e := range ret.Container.Labels {
		p.Labels[e.Name] = e.Value
	}

	return p, nil
}

type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EnvVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func query(ctx context.Context, c *dagger.Client, data interface{}, query string) error {
	return c.Do(
		ctx,
		&dagger.Request{
			Query: query,
		},
		&dagger.Response{
			Data: data,
		},
	)
}
