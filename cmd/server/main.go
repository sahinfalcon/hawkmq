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
	port       = flag.Int("port", 8080, "Server port")
	dataDir    = flag.String("data-dir", "./data", "Data directory")
	persistent = flag.Bool("persistent", true, "Enable queue persistence")
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

	if *persistent {
		if err := os.MkdirAll(*dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}

		logPath := filepath.Join(*dataDir, "queue.log")
		if _, err := os.Stat(logPath); err == nil {
			log.Printf("Found existing log file, recovering queue state...")
			if err := q.RecoverFromLog(logPath); err != nil {
				log.Fatalf("Failure recovering from log: %v", err)
			}
			log.Printf("Successfully recovered queue with %d messages", q.Size())
		} else {
			log.Printf("No existing log found, starting fresh queue with logging")
			if err := q.EnableLogging(logPath); err != nil {
				log.Fatalf("Failed to enable logging: %v", err)
			}
		}
		if err := q.EnableLogging(logPath); err != nil {
			log.Fatalf("Failed to enable logging: %v", err)
		}
		log.Printf("Queue persistence enabled, operations will be logged to %s", logPath)
	} else {
		log.Printf("Queue persistence disabled, queue will be in-memory only")
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
