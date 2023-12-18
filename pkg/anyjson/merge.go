package anyjson

import (
	"context"
)

type MergeOption = func(*merger)

type NullOp int

const (
	NullIgnore NullOp = iota
	NullAsRemover
)

func WithNullOp(op NullOp) MergeOption {
	return func(m *merger) {
		m.nullOp = op
	}
}

func WithArrayMergeKey(key string) MergeOption {
	return func(m *merger) {
		m.arrayMergeKey = key
	}
}

func Merge[T Valuer](base T, patch T, optFns ...MergeOption) T {
	m := &merger{}
	m.Build(optFns...)

	ctx := context.Background()

	switch x := any(patch).(type) {
	case *Object:
		if b, ok := any(base).(*Object); ok {
			return any(m.mergeObject(ctx, b, x)).(T)
		} else {
			return patch
		}
	case *Array:
		if b, ok := any(base).(*Array); ok {
			return any(m.mergeArray(ctx, b, x)).(T)
		} else {
			return patch
		}
	default:
		return patch
	}
}

type merger struct {
	nullOp        NullOp
	arrayMergeKey string
}

func (m *merger) Build(optFns ...MergeOption) {
	for _, fn := range optFns {
		fn(m)
	}
}

func (m *merger) mergeObject(ctx context.Context, left *Object, right *Object) *Object {
	if right == nil {
		return left
	}

	merged := &Object{}

	for leftProp := range left.Iter(ctx) {
		key := leftProp.Entry().(string)
		value := leftProp.Value()

		if rightValue, ok := right.Get(key); ok {
			switch x := rightValue.(type) {
			case *Array:
				if leftValue, ok := value.(*Array); ok {
					value = m.mergeArray(ctx, leftValue, x)
				} else {
					value = x
				}
			case *Object:
				if leftValue, ok := value.(*Object); ok {
					value = m.mergeObject(ctx, leftValue, x)
				} else {
					value = x
				}
			case *Null:
				if m.nullOp == NullIgnore {
					value = &Null{}
					// don't merger null value
					continue
				}
			default:
				value = x
			}
		}

		if _, ok := value.(*Null); !ok {
			merged.Set(key, value)
		}
	}

	for e := range right.Iter(ctx) {
		key := e.Entry().(string)
		value := e.Value()

		switch value.(type) {
		case *Null:
			if m.nullOp == NullIgnore {
				continue
			} else if m.nullOp == NullAsRemover {
				merged.Delete(key)
				continue
			}
		}

		if _, ok := left.Get(key); !ok {
			merged.Set(key, value)
		}
	}

	return merged
}

func (m *merger) mergeArray(ctx context.Context, left *Array, right *Array) *Array {
	if arrayMergeKey := m.arrayMergeKey; arrayMergeKey != "" {
		mergedArr := &Array{}
		processed := map[int]bool{}

		findRightItemObjByValue := func(leftItemMergeKeyValue Valuer) Elem {
			for item := range right.Iter(ctx) {
				if itemObject, ok := item.Value().(*Object); ok {
					if itemMergeKeyValue, ok := itemObject.Get(arrayMergeKey); ok {
						if Equal(itemMergeKeyValue, leftItemMergeKeyValue) {
							return item
						}
					}
				}
			}
			return nil
		}

		for leftItem := range left.Iter(ctx) {
			if leftItemObj, ok := leftItem.Value().(*Object); ok {
				if value, ok := leftItemObj.Get(arrayMergeKey); ok {
					if found := findRightItemObjByValue(value); found != nil {
						idx := found.Entry().(int)
						processed[idx] = true
						mergedArr.Append(m.mergeObject(ctx, leftItemObj, found.Value().(*Object)))
						continue
					}
				}
			}

			mergedArr.Append(leftItem.Value())
		}

		for item := range right.Iter(ctx) {
			idx := item.Entry().(int)
			if _, ok := processed[idx]; ok {
				continue
			}
			mergedArr.Append(item.Value())
		}

		return mergedArr
	}

	return right
}
