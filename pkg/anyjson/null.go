package anyjson

type Null struct {
}

func (*Null) Value() any {
	return nil
}

func (v *Null) String() string {
	return ToString(v)
}

func (v *Null) UnmarshalJSON(bytes []byte) error {
	return nil
}

func (v *Null) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}
