package client

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Env{})
}

// Env of client
type Env struct {
	cueflow.TaskImpl

	// pick the requested env vars
	Env map[string]SecretOrString `json:",inline" output:"env"`
}

func (ce *Env) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ce.Env); err != nil {
		return err
	}

	for k := range ce.Env {
		if strings.HasPrefix(k, "$$") {
			delete(ce.Env, k)
		}
	}

	return nil
}

func (ce *Env) MarshalJSON() ([]byte, error) {

	return json.Marshal(ce.Env)
}

func (ce *Env) Do(ctx context.Context) error {
	secretStore := task.SecretContext.From(ctx)

	clientEnvs := getClientEnvs()

	env := map[string]SecretOrString{}

	for key := range ce.Env {
		e := ce.Env[key]

		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(task.Secret{
					Key:   key,
					Value: envVar,
				})

				env[key] = SecretOrString{
					Secret: SecretOfID(id),
				}
			} else {
				env[key] = SecretOrString{
					Value: envVar,
				}
			}
		} else {
			if secret := e.Secret; secret != nil {
				return errors.Errorf("EnvVar %s is not defined.", key)
			}

			env[key] = e
		}
	}

	ce.Env = env

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
