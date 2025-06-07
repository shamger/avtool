package utils

type Queue struct {
	items []interface{}
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]interface{}, 0),
	}
}

func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) Size() int {
	return len(q.items)
}

func (q *Queue) Peek() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}

func (q *Queue) PeekBack() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	return q.items[len(q.items)-1]
}
