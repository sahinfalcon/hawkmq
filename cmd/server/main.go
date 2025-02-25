// cmd/server/main.go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sahinfalcon/hawkmq/internal/queue"
)

var (
	port     = flag.Int("port", 8080, "Server port")
	dataDir  = flag.String("data-dir", "./data", "Data directory")
	logQueue = flag.Bool("log", false, "Enable queue operation logging")
)

var q = queue.NewQueue()

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := scanner.Text()
		if strings.HasPrefix(msg, "PUBLISH ") {
			q.Enqueue(strings.TrimPrefix(msg, "PUBLISH "))
			fmt.Fprintln(conn, "ACK")
		} else if msg == "CONSUME" {
			if message, ok := q.Dequeue(); ok {
				fmt.Fprintln(conn, message)
			} else {
				fmt.Fprintln(conn, "EMPTY")
			}
		} else if msg == "SIZE" {
			fmt.Fprintf(conn, "SIZE %d\n", q.Size())
		} else {
			fmt.Fprintln(conn, "ERROR: unknown command")
		}
	}
}

func main() {
	flag.Parse()

	if *logQueue {
		if err := os.MkdirAll(*dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}

		logPath := filepath.Join(*dataDir, "queue.log")
		if err := q.EnableLogging(logPath); err != nil {
			log.Fatalf("Failed to enable logging: %v", err)
		}
		log.Printf("Queue operations will be logged to %s", logPath)
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("HawkMQ server listening on port %d", *port)
	log.Printf("Press Ctrl+C to shutdown")

	go func() {
		for {
			select {
			case <-stop:
				close(done)
				return
			default:
				ln.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
				conn, err := ln.Accept()
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					log.Printf("Error accepting connection: %v", err)
					continue
				}

				go handleConnection(conn)
			}
		}
	}()

	<-signalChan
	log.Println("Shutting down...")

	close(stop)
	ln.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Println("Forced shutdown after timeout")
	}

	if err := q.Close(); err != nil {
		log.Printf("Error closing queue: %v", err)
	}

	log.Println("Server stopped")
	os.Exit(0)
}
