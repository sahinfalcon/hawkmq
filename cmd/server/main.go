package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/sahinfalcon/hawkmq/internal/queue"
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
		}
	}
}

func main() {
	ln, _ := net.Listen("tcp", ":8080")
	defer ln.Close()
	fmt.Println("hawkmq server listening on port 8080")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}
