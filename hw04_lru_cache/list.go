package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
	List  // Remove me after realization.
	// Place your code here.
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushBack(v interface{}) *ListItem {
	newBack := &ListItem{
		Value: v,
	}

	switch {
	case l.back != nil:
		l.back.Next = newBack
		newBack.Prev, l.back = l.back, newBack
	case l.front != nil:
		l.front.Next = newBack
		newBack.Prev, l.back = l.front, newBack
	default:
		l.back = newBack
	}
	l.len++
	return l.back.Prev // return old back ?
}

func (l *list) PushFront(v interface{}) *ListItem {
	newFront := &ListItem{
		Value: v,
	}

	switch {
	case l.front != nil:
		l.front.Prev = newFront
		newFront.Next, l.front = l.front, newFront
	case l.back != nil:
		l.front = newFront
		newFront.Next, l.front = l.back, newFront
	default:
		l.front = newFront
	}
	l.len++
	return l.front.Next // return old front ?
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	}

	i.Prev = nil
	i.Next = nil
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || i == l.front {
		return
	}
	l.Remove(i)
	l.PushFront(i.Value)
}

func NewList() List {
	return new(list)
}
