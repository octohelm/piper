package dagger

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

type ImageConfig struct {
	WorkingDir string            `json:"workdir,omitzero"`
	Env        map[string]string `json:"env,omitzero"`
	Labels     map[string]string `json:"label,omitzero"`
	Entrypoint []string          `json:"entrypoint,omitzero"`
	Cmd        []string          `json:"cmd,omitzero"`
	User       string            `json:"user,omitzero"`
}

func ResolveImageConfig(ctx context.Context, c *Client, id dagger.ContainerID) (ImageConfig, error) {
	ret := struct {
		Container struct {
			ID dagger.ContainerID `json:"id,omitzero"`

			Platform string `json:"platform,omitzero"`

			Entrypoint  []string `json:"entrypoint,omitzero"`
			DefaultArgs []string `json:"defaultArgs,omitzero"`
			Workdir     string   `json:"workdir,omitzero"`

			User         string        `json:"user,omitzero"`
			EnvVariables []EnvVariable `json:"envVariables,omitzero"`
			Labels       []Label       `json:"labels,omitzero"`
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
