package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

type Entry struct {
	Value string
	ExpireTime int64
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	//
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	db := make(map[string]Entry)
	var expireTime int64 = -1

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		raw := string(buf[:n])
		args := parseMessage(raw)
		fmt.Printf("%#v\n", args)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "ping":
			conn.Write([]byte("+PONG\r\n"))
		case "echo":
			if len(args) > 1 {
				conn.Write([]byte(respEncoder(args[1])))
			}
		case "set":
			if(args[3] == "ex") {
				expireTime, _ = strconv.ParseInt(args[4][1:], 10, 64)
				expireTime *= 1000
				nowMs := time.Now().UnixMilli()
				expireTime = expireTime + nowMs
			} 
			if(args[3] == "px"){
				expireTime, _ = strconv.ParseInt(args[4][1:], 10, 64)
				nowMs := time.Now().UnixMilli()
				expireTime = expireTime + nowMs
			}
			
			db[args[1]] = Entry{Value: args[2], ExpireTime: expireTime}
			conn.Write([]byte(setValue()))
		case "get":
			conn.Write([]byte(getValue(db[args[1]])))
		}
			
	}
}

func parseMessage(message string) []string {
	lines := strings.Split(message, "\r\n")
	var args []string

	i := 0
	if len(lines) == 0 {
		return args
	}

	// First line should be *N (array with N elements)
	if len(lines[0]) == 0 || lines[0][0] != '*' {
		return args
	}
	numArgs, err := strconv.Atoi(lines[0][1:])
	if err != nil {
		return args
	}
	i++

	for j := 0; j < numArgs; j++ {
		if i >= len(lines) {
			break
		}
		// Skip the $N bulk string length prefix
		if len(lines[i]) > 0 && lines[i][0] == '$' {
			i++
		}
		if i >= len(lines) {
			break
		}
		args = append(args, strings.ToLower(lines[i]))
		i++
	}

	return args
}

func respEncoder(raw string) string {
	// $<length>\r\n<data>\r\n
	return fmt.Sprintf("$%d\r\n%s\r\n", len(raw), raw)
}

func simpleEncoder(raw string) string {
	return fmt.Sprintf("+%s\r\n", raw)
}

func setValue() string {
	return simpleEncoder("OK")
}

func getValue(entry Entry) string {
	if (entry.Value == "") {
		return "$-1\r\n"
	}
	
	if (entry.ExpireTime == -1) {
		return  respEncoder(entry.Value)
	}
	
	nowMs := time.Now().UnixMilli()
	
	if(nowMs > entry.ExpireTime) {
		return  respEncoder(entry.Value)
	}
	
	return "$-1\r\n"
}