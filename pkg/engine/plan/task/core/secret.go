package core

import (
	"context"
	"encoding/json"
	"github.com/octohelm/piper/pkg/engine/rigutil"
)

func init() {
	DefaultFactory.Register(&Secret{})
}

func SecretOfID(id string) *Secret {
	s := &Secret{}
	s.Meta.Secret.ID = id
	return s
}

type Secret struct {
	Meta struct {
		Secret struct {
			ID string `json:"id,omitempty"`
		} `json:"secret"`
	} `json:"$piper"`
}

func (s *Secret) Value(ctx context.Context) (rigutil.Secret, bool) {
	return rigutil.SecretContext.From(ctx).Get(s.Meta.Secret.ID)
}

type SecretOrString struct {
	Value  string
	Secret *Secret
}

func (SecretOrString) OneOf() []any {
	return []any{
		"",
		&Secret{},
	}
}

func (s *SecretOrString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '{' {
		se := &Secret{}
		if err := json.Unmarshal(data, se); err != nil {
			return err
		}
		s.Secret = se
		return nil
	}
	return json.Unmarshal(data, &s.Value)
}

func (s SecretOrString) MarshalJSON() ([]byte, error) {
	if s.Secret != nil {
		return json.Marshal(s.Secret)
	}
	return json.Marshal(s.Value)
}
