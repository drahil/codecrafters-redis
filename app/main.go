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
	Value      string
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
	lists := make(map[string][]string)

	for {
		args, err := getArgs(conn)

		if err != nil {
			break
		}

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
			conn.Write([]byte(setValue(args, db)))
		case "get":
			conn.Write([]byte(getValue(db[args[1]])))
		case "rpush":
			conn.Write([]byte(rpushValue(args, lists)))
		case "lrange":
			conn.Write([]byte(lrange(args, lists)))
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

func getArgs(conn net.Conn) ([]string, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	raw := string(buf[:n])
	args := parseMessage(raw)
	fmt.Printf("%#v\n", args)

	return args, err
}

func respEncoder(raw string) string {
	// $<length>\r\n<data>\r\n
	return fmt.Sprintf("$%d\r\n%s\r\n", len(raw), raw)
}

func simpleEncoder(raw string) string {
	return fmt.Sprintf("+%s\r\n", raw)
}

func respInteger(raw int) string {
	stringRaw := strconv.Itoa(raw)
	return fmt.Sprintf(":%s\r\n", stringRaw)
}

func respArray(values []string) string {
	if len(values) == 0 {
		return "*0\r\n"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(values)))

	for _, value := range values {
		builder.WriteString(respEncoder(value))
	}

	return builder.String()
}

func setValue(args []string, db map[string]Entry) string {
	var expireTime int64 = -1

	if len(args) > 3 && args[3] == "ex" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		expireTime *= 1000
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}
	if len(args) > 4 && args[3] == "px" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}

	db[args[1]] = Entry{Value: args[2], ExpireTime: expireTime}
	return simpleEncoder("OK")
}

func getValue(entry Entry) string {
	if entry.Value == "" {
		return "$-1\r\n"
	}

	if entry.ExpireTime == -1 {
		return respEncoder(entry.Value)
	}

	nowMs := time.Now().UnixMilli()

	if nowMs <= entry.ExpireTime {
		return respEncoder(entry.Value)
	}

	return "$-1\r\n"
}

func rpushValue(args []string, lists map[string][]string) string {
	listName := args[1]
	values := args[2:]

	if existingList, ok := lists[listName]; ok {
		lists[listName] = append(existingList, values...)
	} else {
		lists[listName] = values
	}

	return respInteger(len(lists[listName]))
}

func lrange(args []string, lists map[string][]string) string {
	listName := args[1]
	startingIndex, _ := strconv.ParseInt(args[2], 10, 64)
	endingIndex, _ := strconv.ParseInt(args[3], 10, 64)
	
	if existingList, ok := lists[listName]; ok {
		if startingIndex >= int64(len(existingList)) {
			return respArray([]string{})
		}
		
		if endingIndex >= int64(len(existingList)) {
			return respArray(existingList[startingIndex:len(existingList)-1])
		}
		
		if startingIndex > endingIndex {
			return respArray([]string{})
		}
		return respArray(existingList[startingIndex:endingIndex])
	}
	
	return respArray([]string{})
}
