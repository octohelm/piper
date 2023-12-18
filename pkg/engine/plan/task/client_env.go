package task

import (
	"context"
	"encoding/json"
	"github.com/octohelm/piper/pkg/engine/rigutil"
	"github.com/pkg/errors"
	"os"
	"strings"
	"sync"

	"github.com/octohelm/piper/pkg/engine/plan/task/core"
)

func init() {
	core.DefaultFactory.Register(&ClientEnv{})
}

type ClientEnv struct {
	core.Task
	Env map[string]core.SecretOrString `json:",inline" piper:"generated"`
}

func (ce *ClientEnv) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &ce.Env); err != nil {
		return err
	}
	delete(ce.Env, "$piper")
	return nil
}

func (ce *ClientEnv) MarshalJSON() ([]byte, error) {
	return json.Marshal(ce.Env)
}

func (ce *ClientEnv) Do(ctx context.Context) error {
	secretStore := rigutil.SecretContext.From(ctx)

	clientEnvs := getClientEnvs()

	env := map[string]core.SecretOrString{}

	for key := range ce.Env {
		e := ce.Env[key]

		if envVar, ok := clientEnvs[key]; ok {
			if secret := e.Secret; secret != nil {
				id := secretStore.Set(rigutil.Secret{
					Key:   key,
					Value: envVar,
				})

				env[key] = core.SecretOrString{
					Secret: core.SecretOfID(id),
				}
			} else {
				env[key] = core.SecretOrString{
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
