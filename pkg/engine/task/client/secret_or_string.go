package client

import (
	"encoding/json"
)

type SecretOrString struct {
	Secret *Secret
	Value  string
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
