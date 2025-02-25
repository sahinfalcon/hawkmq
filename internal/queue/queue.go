package queue

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Queue struct {
	messages      []string
	mu            sync.RWMutex
	logFile       *os.File
	enableLogging bool
}

func NewQueue() *Queue {
	return &Queue{messages: []string{}, enableLogging: false}
}

func (q *Queue) EnableLogging(logPath string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	dir := logPath[:len(logPath)-len("/queue.log")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failure in creating log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failure in opening log file: %w", err)
	}

	q.logFile = file
	q.enableLogging = true

	for _, msg := range q.messages {
		fmt.Fprintf(q.logFile, "ENQUEUE %s %s\n", time.Now().Format(time.RFC3339), msg)
	}
	return nil
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

func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.logFile != nil {
		return q.logFile.Close()
	}
	return nil
}
