package queue

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Queue struct {
	messages      []string
	mu            sync.RWMutex
	logFile       *os.File
	enableLogging bool
	logPath       string
}

func NewQueue() *Queue {
	return &Queue{
		messages:      []string{},
		enableLogging: false,
	}
}

func (q *Queue) EnableLogging(logPath string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.logPath = logPath

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failure in creating log directory: %w", err)
	}

	fmt.Printf("Opening log file at: %s\n", logPath)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failure in opening log file: %w", err)
	}

	q.logFile = file
	q.enableLogging = true

	timestamp := time.Now().Format(time.RFC3339)
	startupMsg := fmt.Sprintf("STARTUP %s Queue logging enabled\n", timestamp)
	if _, err := q.logFile.WriteString(startupMsg); err != nil {
		return fmt.Errorf("failed to write startup message: %w", err)
	}

	if err := q.logFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync log file: %w", err)
	}

	fmt.Printf("Logging enabled successfully\n")

	for _, msg := range q.messages {
		logMsg := fmt.Sprintf("ENQUEUE %s %s\n", time.Now().Format(time.RFC3339), msg)
		if _, err := q.logFile.WriteString(logMsg); err != nil {
			return fmt.Errorf("failed to log existing message: %w", err)
		}
	}

	return q.logFile.Sync()
}

func (q *Queue) RecoverFromLog(logPath string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.logPath = logPath

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failure opening log file for recovery: %w", err)
	}
	defer file.Close()

	q.messages = []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}

		operation := parts[0]
		//timestamp := parts[1]
		message := parts[2]

		switch operation {
		case "ENQUEUE":
			q.messages = append(q.messages, message)
		case "DEQUEUE":
			for i, msg := range q.messages {
				if msg == message {
					q.messages = append(q.messages[:i], q.messages[i+1:]...)
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	appendFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failure reopening log file for appending: %w", err)
	}

	q.logFile = appendFile
	q.enableLogging = true

	return nil
}

func (q *Queue) Enqueue(msg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.messages = append(q.messages, msg)

	if q.enableLogging && q.logFile != nil {
		logMsg := fmt.Sprintf("ENQUEUE %s %s\n", time.Now().Format(time.RFC3339), msg)
		_, err := q.logFile.WriteString(logMsg)
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
			return
		}

		if err := q.logFile.Sync(); err != nil {
			fmt.Printf("Error syncing log file: %v\n", err)
		}
	}
}

func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.messages) == 0 {
		return "", false
	}
	msg := q.messages[0]
	q.messages = q.messages[1:]

	if q.enableLogging && q.logFile != nil {
		logMsg := fmt.Sprintf("DEQUEUE %s %s\n", time.Now().Format(time.RFC3339), msg)
		_, err := q.logFile.WriteString(logMsg)
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
			return msg, true
		}

		if err := q.logFile.Sync(); err != nil {
			fmt.Printf("Error syncing log file: %v\n", err)
		}
	}

	return msg, true
}

func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.messages)
}

func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.logFile != nil {
		timestamp := time.Now().Format(time.RFC3339)
		shutdownMsg := fmt.Sprintf("SHUTDOWN %s Queue closed\n", timestamp)
		q.logFile.WriteString(shutdownMsg)
		q.logFile.Sync()

		return q.logFile.Close()
	}
	return nil
}
