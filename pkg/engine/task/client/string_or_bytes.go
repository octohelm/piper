package client

import (
	"encoding/json"
)

type StringOrBytes []byte

func (StringOrBytes) OneOf() []any {
	return []any{
		"",
		[]byte{},
	}
}

func (s *StringOrBytes) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		b := ""
		if err := json.Unmarshal(data, &b); err != nil {
			return err
		}
		*s = []byte(b)

		return nil
	}
	*s = data
	return nil
}

func (s StringOrBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(s))
}
