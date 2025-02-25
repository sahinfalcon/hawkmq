package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")
	defer conn.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <publish|consume> [message]")
		return
	}

	action := os.Args[1]
	if action == "publish" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go publish <message>")
			return
		}
		fmt.Fprintln(conn, "PUBLISH "+os.Args[2])
	} else if action == "consume" {
		fmt.Fprintln(conn, "CONSUME")
	} else if action == "size" {
		fmt.Fprintln(conn, "SIZE")
	}

	response, _ := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Response:", response)
}
