package anyjson

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/octohelm/piper/pkg/generic/list"
	"github.com/pkg/errors"
)

type Object struct {
	props map[string]*list.Node[*field]
	ll    list.List[*field]
}

type field struct {
	key   string
	value Valuer
}

func (f *field) Entry() any {
	return f.key
}

func (f *field) Value() Valuer {
	return f.value
}

func (f *field) Set(v Valuer) {
	f.value = v
}

func (v *Object) Value() any {
	m := map[string]any{}
	for k, e := range v.props {
		m[k] = e.Value.Value().Value()
	}
	return m
}

func (v *Object) Len() int {
	return len(v.props)
}

func (v *Object) Iter(ctx context.Context) <-chan Elem {
	if v == nil || v.ll.Front == nil {
		return closedElementCh
	}

	ch := make(chan Elem)

	go func() {
		defer func() {
			close(ch)
		}()

		for el := v.ll.Front; el != nil; el = el.Next {
			select {
			case <-ctx.Done():
			case ch <- el.Value:
			}
		}
	}()

	return ch
}

func (v Object) Get(key string) (Valuer, bool) {
	if v.props != nil {
		v, ok := v.props[key]
		if ok {
			return v.Value.Value(), true
		}
	}
	return nil, false
}

func (v *Object) Set(key string, value Valuer) bool {
	if v.props == nil {
		v.props = map[string]*list.Node[*field]{}
	}

	_, alreadyExist := v.props[key]
	if alreadyExist {
		v.props[key].Value.Set(value)
		return false
	}

	element := &field{key: key, value: value}
	v.props[key] = v.ll.PushBack(element)
	return true
}

func (v *Object) Delete(key string) (didDelete bool) {
	if v.props == nil {
		return false
	}

	element, ok := v.props[key]
	if ok {
		v.ll.Remove(element)

		delete(v.props, key)
	}
	return ok
}

func (v *Object) UnmarshalJSON(b []byte) error {
	if v == nil {
		*v = Object{}
	}

	d := jsontext.NewDecoder(bytes.NewReader(b))

	t, err := d.ReadToken()
	if err != nil {
		return err
	}

	if t.Kind() != '{' {
		return errors.New("object should starts with `{`")
	}

	for kind := d.PeekKind(); kind != '}'; kind = d.PeekKind() {
		k, err := d.ReadValue()
		if err != nil {
			return err
		}

		key, err := strconv.Unquote(string(k))
		if err != nil {
			return err
		}

		var value Valuer
		if err := UnmarshalTo(d, &value); err != nil {
			return err
		}

		v.Set(key, value)
	}

	return nil
}

func (v *Object) MarshalJSON() ([]byte, error) {
	b := bytes.NewBuffer(nil)

	b.WriteString("{")

	idx := 0
	for e := range v.Iter(context.Background()) {
		if idx > 0 {
			b.WriteString(",")
		}

		b.WriteString(strconv.Quote(fmt.Sprint(e.Entry().(string))))
		b.WriteString(":")
		raw, err := e.Value().MarshalJSON()
		if err != nil {
			return []byte{}, err
		}
		b.Write(raw)
		idx++
	}

	b.WriteString("}")

	return b.Bytes(), nil
}

func (v *Object) String() string {
	return ToString(v)
}
