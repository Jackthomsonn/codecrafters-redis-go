package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/jackthomsonn/redis-go-impl/internal"
	"github.com/jackthomsonn/redis-go-impl/store"
)

func main() {
	redisStore := store.NewRedisStore()

	go store.RunRemovalCheck(redisStore)

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379:", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn, redisStore)
	}
}

func parseRedisCommand(readBytes []byte) (string, internal.ParsedResponse, error) {
	var res internal.ParsedResponse
	bytesAsString := string(readBytes)
	components := strings.Split(bytesAsString, "\r\n")

	if len(components) < 3 {
		return "", res, fmt.Errorf("invalid command format")
	}

	command := strings.ToLower(components[2])

	if len(components) > 10 {
		res.Px = components[8]
		mili, err := strconv.ParseInt(components[10], 10, 64)
		if err != nil {
			return "", res, fmt.Errorf("error parsing milli: %w", err)
		}
		res.Mili = mili
	}

	if len(components) > 4 {
		res.Key = components[4]
	}

	if len(components) > 6 {
		res.Value = components[6]
	}

	return command, res, nil
}

func encodeBulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func handleConnection(conn net.Conn, redisStore *store.RedisStore) {
	defer conn.Close()
	buf := make([]byte, 1024)

	for {
		size, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		command, parsedResponse, err := parseRedisCommand(buf[:size])
		if err != nil {
			fmt.Println("Error parsing command:", err)
			return
		}

		switch command {
		case "ping":
			if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
				fmt.Println("Error sending PONG:", err)
				return
			}
		case "echo":
			if _, err := conn.Write([]byte(encodeBulkString(parsedResponse.Key))); err != nil {
				fmt.Println("Error sending ECHO:", err)
				return
			}
		case "set":
			redisStore.Set(parsedResponse)
			if _, err := conn.Write([]byte("+OK\r\n")); err != nil {
				fmt.Println("Error sending OK:", err)
				return
			}
		case "get":
			val, err := redisStore.Get(parsedResponse.Key)
			if err != nil {
				if _, err := conn.Write([]byte("$-1\r\n")); err != nil {
					fmt.Println("Error sending nil response:", err)
					return
				}
			} else {
				if _, err := conn.Write([]byte(encodeBulkString(val.(string)))); err != nil {
					fmt.Println("Error sending GET response:", err)
					return
				}
			}
		default:
			fmt.Println("Received unknown command:", command)
		}
	}
}
