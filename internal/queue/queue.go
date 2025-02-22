package queue

import "sync"

type Queue struct {
	messages []string
	mu       sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{messages: []string{}}
}

func (q *Queue) Enqueue(msg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.messages) == 0 {
		return "", false
	}
	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg, true
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}
