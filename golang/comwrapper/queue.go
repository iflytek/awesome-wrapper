package comwrapper

import (
	"sync"
)

// Queue 是一个并发安全的队列结构
type Queue struct {
	items []interface{}
	lock  sync.Mutex
}

// NewQueue 创建一个新的队列
func NewQueue() *Queue {
	return &Queue{
		items: make([]interface{}, 0),
		lock:  sync.Mutex{},
	}
}

// Enqueue 将元素放入队列尾部
func (q *Queue) Enqueue(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.items = append(q.items, item)
}

// Dequeue 从队列头部取出元素
func (q *Queue) Dequeue() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// Len 返回队列中元素的个数
func (q *Queue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.items)
}
