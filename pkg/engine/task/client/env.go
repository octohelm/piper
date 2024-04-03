package client

import (
	"context"
	"os"
	"strings"
	"sync"

	"cuelang.org/go/cue"

	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &EnvInterface{})
}

// EnvInterface of client
type EnvInterface struct {
	// to avoid added ok
	task.Task `json:"-"`

	RequiredEnv map[string]SecretOrString `json:"-"`
	OptionalEnv map[string]SecretOrString `json:"-"`
}

func (EnvInterface) CacheDisabled() bool {
	return true
}

var _ cueflow.CacheDisabler = &EnvInterface{}

func (ei *EnvInterface) UnmarshalTask(t cueflow.Task) error {
	v := cueflow.CueValue(t.Value())

	i, err := v.Fields(cue.All())
	if err != nil {
		return err
	}

	ei.RequiredEnv = make(map[string]SecretOrString)
	ei.OptionalEnv = make(map[string]SecretOrString)

	for i.Next() {
		envKey := i.Selector().Unquoted()

		if strings.HasPrefix(envKey, "$$") {
			continue
		}

		v := SecretOrString{}

		if i.Value().LookupPath(SecretPath).Exists() {
			v.Secret = &Secret{}
		}

		if i.Selector().Type()&cue.RequiredConstraint != 0 {
			ei.RequiredEnv[envKey] = v
		} else {
			ei.OptionalEnv[envKey] = v
		}
	}

	return nil
}

var _ cueflow.TaskUnmarshaler = &EnvInterface{}

var _ cueflow.OutputValuer = EnvInterface{}

func (ei EnvInterface) OutputValues() map[string]any {
	values := map[string]any{}

	for k, v := range ei.RequiredEnv {
		values[k] = v
	}

	for k, v := range ei.OptionalEnv {
		values[k] = v
	}

	return values
}

func (ei *EnvInterface) Do(ctx context.Context) error {
	secretStore := task.SecretContext.From(ctx)
	clientEnvs := getClientEnvs()

	for key, e := range ei.RequiredEnv {
		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(task.Secret{
					Key:   key,
					Value: envVar,
				})
				ei.RequiredEnv[key] = SecretOrString{
					Secret: SecretOfID(id),
				}
			} else {
				ei.RequiredEnv[key] = SecretOrString{
					Value: envVar,
				}
			}
		} else {
			return errors.Errorf("EnvVar %s is required, but not defined.", key)
		}
	}

	for key, e := range ei.OptionalEnv {
		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(task.Secret{
					Key:   key,
					Value: envVar,
				})
				ei.OptionalEnv[key] = SecretOrString{
					Secret: SecretOfID(id),
				}
			} else {
				ei.OptionalEnv[key] = SecretOrString{
					Value: envVar,
				}
			}
		}
	}

	return nil
}

var getClientEnvs = sync.OnceValue(func() map[string]string {
	clientEnvs := map[string]string{}

	for _, i := range os.Environ() {
		parts := strings.SplitN(i, "=", 2)
		if len(parts) == 2 {
			clientEnvs[parts[0]] = parts[1]
		} else {
			clientEnvs[parts[0]] = ""
		}
	}

	return clientEnvs
})
