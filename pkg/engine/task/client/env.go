package client

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"github.com/octohelm/cuekit/pkg/cueconvert"
	"github.com/octohelm/cuekit/pkg/cueflow"
	"github.com/octohelm/cuekit/pkg/cueflow/task"
	enginetask "github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	enginetask.Registry.Register(&EnvInterface{})
}

// EnvInterface of client
type EnvInterface struct {
	// to avoid added ok
	task.Task `json:"-"`

	RequiredEnv map[string]SecretOrString `json:"-"`
	OptionalEnv map[string]SecretOrString `json:"-"`
}

var _ cueflow.CacheDisabler = &EnvInterface{}

func (EnvInterface) CacheDisabled() bool {
	return true
}

var _ cueflow.CueValueUnmarshaler = &EnvInterface{}

func (ei *EnvInterface) UnmarshalCueValue(v cue.Value) error {
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

		if i.Value().Kind() == cue.StructKind {
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

var _ cueconvert.OutputValuer = EnvInterface{}

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
	secretStore := enginetask.SecretContext.From(ctx)
	clientEnvs := getClientEnvs()

	for key, e := range ei.RequiredEnv {
		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(enginetask.Secret{
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
			return fmt.Errorf("env var %s is required, but not defined", key)
		}
	}

	for key, e := range ei.OptionalEnv {
		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(enginetask.Secret{
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
