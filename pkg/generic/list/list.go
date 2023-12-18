package list

type List[V any] struct {
	Front, Back *Node[V]
}

type Node[V any] struct {
	Value      V
	Prev, Next *Node[V]
}

func (l *List[V]) PushBack(v V) *Node[V] {
	n := &Node[V]{
		Value: v,
	}
	l.PushBackNode(n)
	return n
}

func (l *List[V]) PushFront(v V) *Node[V] {
	n := &Node[V]{
		Value: v,
	}
	l.PushFrontNode(n)
	return n
}

func (l *List[V]) PushBackNode(n *Node[V]) {
	n.Next = nil
	n.Prev = l.Back
	if l.Back != nil {
		l.Back.Next = n
	} else {
		l.Front = n
	}
	l.Back = n
}

func (l *List[V]) PushFrontNode(n *Node[V]) {
	n.Next = l.Front
	n.Prev = nil
	if l.Front != nil {
		l.Front.Prev = n
	} else {
		l.Back = n
	}
	l.Front = n
}

func (l *List[V]) Remove(n *Node[V]) {
	if n.Next != nil {
		n.Next.Prev = n.Prev
	} else {
		l.Back = n.Prev
	}
	if n.Prev != nil {
		n.Prev.Next = n.Next
	} else {
		l.Front = n.Next
	}
}
