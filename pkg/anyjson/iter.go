package anyjson

import "context"

var closedElementCh = make(chan Elem)

func init() {
	close(closedElementCh)
}

type ElemIter interface {
	Iter(ctx context.Context) <-chan Elem
}

type Elem interface {
	Entry() any
	Value() Valuer
}
