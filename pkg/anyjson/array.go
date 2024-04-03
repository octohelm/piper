package anyjson

import (
	"bytes"
	"context"

	"github.com/go-json-experiment/json/jsontext"
	"github.com/pkg/errors"
)

type Array struct {
	items []Valuer
}

func (v *Array) Value() any {
	list := make([]any, len(v.items))
	for i := range list {
		list[i] = v.items[i].Value()
	}
	return list
}

func (v *Array) Len() int {
	return len(v.items)
}

func (v *Array) Iter(ctx context.Context) <-chan Elem {
	if len(v.items) == 0 {
		return closedElementCh
	}

	ch := make(chan Elem)

	go func() {
		defer func() {
			close(ch)
		}()

		for i, v := range v.items {
			select {
			case <-ctx.Done():
			case ch <- &arrItem{idx: i, value: v}:
			}
		}
	}()

	return ch
}

type arrItem struct {
	idx   int
	value Valuer
}

func (i *arrItem) Entry() any {
	return i.idx
}

func (i *arrItem) Value() Valuer {
	return i.value
}

func (v *Array) UnmarshalJSON(b []byte) error {
	if v == nil {
		*v = Array{}
	}

	d := jsontext.NewDecoder(bytes.NewReader(b))

	t, err := d.ReadToken()
	if err != nil {
		return err
	}

	if t.Kind() != '[' {
		return errors.New("v should starts with `[`")
	}

	for kind := d.PeekKind(); kind != ']'; kind = d.PeekKind() {
		var value Valuer
		if err := UnmarshalTo(d, &value); err != nil {
			return nil
		}
		v.items = append(v.items, value)
	}

	return nil
}

func (v *Array) MarshalJSON() ([]byte, error) {
	b := bytes.NewBuffer(nil)

	b.WriteString("[")

	for idx, v := range v.items {
		if idx > 0 {
			b.WriteString(",")
		}

		raw, err := v.MarshalJSON()
		if err != nil {
			return []byte{}, err
		}
		b.Write(raw)
		idx++
	}

	b.WriteString("]")

	return b.Bytes(), nil
}

func (v *Array) String() string {
	return ToString(v)
}

func (v *Array) Append(item Valuer) {
	v.items = append(v.items, item)
}
