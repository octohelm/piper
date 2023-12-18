package task

import (
	"context"
	"github.com/octohelm/piper/pkg/wd"
	"github.com/pkg/errors"
	"strings"

	"github.com/octohelm/piper/pkg/engine/plan/task/core"
)

func init() {
	core.DefaultFactory.Register(&Exec{})
}

type Exec struct {
	core.Task

	CWD core.WD `json:"cwd"`

	Command string                         `json:"cmd"`
	Args    []string                       `json:"args,omitempty"`
	Env     map[string]core.SecretOrString `json:"env,omitempty"`

	User string `json:"user,omitempty"`

	Result core.Result `json:"-" piper:"generated,name=result"`
}

func (e *Exec) Do(ctx context.Context) error {
	return e.CWD.Do(
		ctx,
		func(rootfs wd.WorkDir) error {
			cmd := strings.Join(append([]string{e.Command}, e.Args...), " ")

			env := map[string]string{}

			for k, v := range e.Env {
				if s := v.Secret; s != nil {
					if secret, ok := s.Value(ctx); ok {
						env[k] = secret.Value
					} else {
						return errors.Errorf("not found secret for %s", k)
					}
				} else {
					env[k] = v.Value
				}
			}

			if err := rootfs.Exec(ctx, cmd,
				wd.WithEnv(env),
				wd.WithUser(e.User),
			); err != nil {
				return err
			}

			e.Result.Ok = true

			return nil
		},
	)
}
