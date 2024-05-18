package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/store"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	store := store.NewRedisStore()

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

		go handlePing(conn, store)
	}
}

func parseRedisCommand(readBytes []byte) (string, string, string) {
	bytesAsString := string(readBytes)
	fmt.Println(bytesAsString)
	components := strings.Split(bytesAsString, "\r\n")
	fmt.Println(components)
	command := components[2]
	arg := ""
	val := ""
	if len(components)-1 > 4 {
		arg = components[4]
		if len(components)-1 > 6 {
			val = components[6]
		}
	}
	return strings.ToLower(command), arg, val
}

func encodeBulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func handlePing(conn net.Conn, store *store.RedisStore) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		size, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading message: ", err.Error())
			return
		}

		command, arg, val := parseRedisCommand(buf[:size])

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
		case "set":
			fmt.Println("Received SET command with arg: ", arg)
			store.Set(arg, val)
			_, err = conn.Write([]byte("+OK\r\n"))
			if err != nil {
				fmt.Println("Received error: ", err.Error())
				return
			}
			fmt.Println("Sent OK")
		case "get":
			fmt.Println("Received GET command with arg: ", arg)
			val := store.Get(arg)
			_, err = conn.Write([]byte(encodeBulkString(val)))
			if err != nil {
				fmt.Println("Received error: ", err.Error())
				return
			}
			fmt.Println("Sent GET response")
		default:
			fmt.Println("Received unknown message: ", command)
		}
	}
}
