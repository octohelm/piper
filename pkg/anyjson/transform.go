package anyjson

import "context"

func Transform(ctx context.Context, v Valuer, transform func(v Valuer, keyPath ...any) Valuer) Valuer {
	t := &transformer{
		transform: transform,
	}
	return t.Next(ctx, v, nil)
}

type transformer struct {
	transform func(v Valuer, keyPath ...any) Valuer
}

func (t *transformer) Next(ctx context.Context, v Valuer, keyPath []any) Valuer {
	switch x := v.(type) {
	case *Object:
		o := &Object{}

		for e := range x.Iter(ctx) {
			propValue := t.Next(ctx, e.Value(), append(keyPath, e.Entry()))

			if propValue != nil {
				o.Set(e.Entry().(string), propValue)
			}
		}

		return o
	case *Array:
		a := &Array{}

		for e := range x.Iter(ctx) {
			if itemValue := t.Next(ctx, e.Value(), append(keyPath, e.Entry())); itemValue != nil {
				a.Append(itemValue)
			}
		}

		return a
	default:
		return t.transform(v, keyPath...)
	}
}
