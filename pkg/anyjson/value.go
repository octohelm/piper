package anyjson

import (
	"context"
	"fmt"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/piper/pkg/generic/ordered"
	"github.com/octohelm/x/ptr"
	"github.com/pkg/errors"
)

func Equal(a Valuer, b Valuer) bool {
	return a.Value() == b.Value()
}

func From(value any) Valuer {
	switch x := value.(type) {
	case []any:
		arr := &Array{}
		for _, e := range x {
			arr.Append(From(e))
		}
		return arr
	case map[string]any:
		o := &Object{}
		for f := range ordered.IterMap(context.Background(), x) {
			o.Set(f.Key, From(f.Value))
		}
		return o
	case string:
		return &String{value: &x}
	case bool:
		return &Boolean{value: &x}
	case int:
		return &Number{value: ptr.Ptr(float64(x))}
	case int8:
	case int16:
	case int32:
	case int64:
	}

	return &Null{}
}

type Valuer interface {
	json.MarshalerV1
	fmt.Stringer
	Value() any
}

func ToString(valuer Valuer) string {
	data, _ := valuer.MarshalJSON()
	return string(data)
}

func UnmarshalTo[T *Valuer](decoder *jsontext.Decoder, t T) error {
	value, err := decoder.ReadValue()
	if err != nil {
		return err
	}

	switch value.Kind() {
	case 'n':
		*t = &Null{}
		return nil
	case 'f':
		*t = &Boolean{raw: value.Clone()}
		return nil
	case 't':
		*t = &Boolean{raw: value.Clone()}
		return nil
	case '"':
		*t = &String{raw: value.Clone()}
		return nil
	case '0':
		*t = &Number{raw: value.Clone()}
		return nil
	case '{':
		o := &Object{}
		if err := o.UnmarshalJSON(value); err != nil {
			return err
		}
		*t = o
		return nil
	case '[':
		arr := &Array{}
		if err := arr.UnmarshalJSON(value); err != nil {
			return err
		}
		*t = arr
		return nil
	}

	return errors.New("invalid value")
}
