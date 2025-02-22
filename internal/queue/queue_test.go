package queue

import (
	"strconv"
	"sync"
	"testing"
)

func TestQueue_Basics(t *testing.T) {
	q := NewQueue()

	// test enqueue and dequeue
	q.Enqueue("message1")
	q.Enqueue("message2")

	if q.Size() != 2 {
		t.Errorf("expected size 2, got %d", q.Size())
	}

	msg, ok := q.Dequeue()
	if !ok || msg != "message1" {
		t.Errorf("expected 'message1', got %v", msg)
	}

	msg, ok = q.Dequeue()
	if !ok || msg != "message2" {
		t.Errorf("expected 'message2', got %v", msg)
	}

	// test empty queue case
	msg, ok = q.Dequeue()
	if ok {
		t.Errorf("expected empty queue, got %v", msg)
	}
}

func TestQueue_Concurrency(t *testing.T) {
	q := NewQueue()
	var wg sync.WaitGroup

	// concurrent enqueues
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q.Enqueue(strconv.Itoa(i))
		}(i)
	}
	wg.Wait()

	if q.Size() != 100 {
		t.Errorf("expected size 100, got %d", q.Size())
	}

	// concurrent dequeues
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Dequeue()
		}()
	}
	wg.Wait()

	if q.Size() != 0 {
		t.Errorf("expected empty queue, got size %d", q.Size())
	}
}
