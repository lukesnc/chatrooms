package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	var shouldQuit bool = false

	// Read from server
	go func() {
		for !shouldQuit {
			buf := make([]byte, 256)
			_, err := conn.Read(buf)
			if err == io.EOF {
				shouldQuit = true
			}

			fmt.Print(string(buf))
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Send input
	reader := bufio.NewReader(os.Stdin)
	for !shouldQuit {
		input, _ := reader.ReadString('\n')
		conn.Write([]byte(input))
	}
}
