package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handlePing(conn)
	}
}

func parseRedisCommand(readBytes []byte) (string, string) {
	bytesAsString := string(readBytes)
	fmt.Println(bytesAsString)
	components := strings.Split(bytesAsString, "\r\n")
	fmt.Println(components)
	command := components[2]
	arg := ""
	if len(components)-1 > 4 {
		arg = components[4]
	}
	return strings.ToLower(command), arg
}

func encodeBulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func handlePing(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		size, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading message: ", err.Error())
			return
		}

		command, arg := parseRedisCommand(buf[:size])

		switch command {
		case "ping":
			_, err = conn.Write([]byte("+PONG\r\n"))

			if err != nil {
				fmt.Println("Received error: ", err.Error())
				return
			}
			fmt.Println("Sent PONG")
		case "echo":
			_, err = conn.Write([]byte(encodeBulkString(arg)))
			if err != nil {
				fmt.Println("Received error: ", err.Error())
				return
			}
			fmt.Println("Sent ECHO")
		default:
			fmt.Println("Received unknown message: ", command)
		}
	}
}
