package client

import "encoding/json"

type StringOrSlice []string

func (StringOrSlice) OneOf() []any {
	return []any{
		"",
		[]string{},
	}
}

func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		b := ""
		if err := json.Unmarshal(data, &b); err != nil {
			return err
		}
		*s = []string{b}
		return nil
	}

	var list []string

	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	*s = list

	return nil
}

func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(s))
}
