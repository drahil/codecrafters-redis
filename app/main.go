package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

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
	
	store := store.New()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, store)
	}
}

func handleConnection(conn net.Conn, store *store.Store) {
	defer conn.Close()
	
	for {
		args, err := resp.GetArgs(conn)

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
				conn.Write([]byte(resp.BulkString(args[1])))
			}
		case "set":
			conn.Write([]byte(setValue(args, store)))
		case "get":
			conn.Write([]byte(getValue(store.Strings[args[1]])))
		case "rpush":
			conn.Write([]byte(rpushValue(args, store.Lists)))
		case "lrange":
			conn.Write([]byte(lrange(args, store.Lists)))
		}
			
	}
}

func setValue(args []string, store *store.Store) string {
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
	
	store.Set(args[1], args[2], expireTime)

	return resp.SimpleString("OK")
}


func getValue(entry store.Entry) string {
	if entry.Value == "" {
		return "$-1\r\n"
	}

	if entry.ExpireTime == -1 {
		return resp.BulkString(entry.Value)
	}

	nowMs := time.Now().UnixMilli()

	if nowMs <= entry.ExpireTime {
		return resp.BulkString(entry.Value)
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

	return resp.Integer(len(lists[listName]))
}

func lrange(args []string, lists map[string][]string) string {
	listName := args[1]
	startingIndex, _ := strconv.ParseInt(args[2], 10, 64)
	endingIndex, _ := strconv.ParseInt(args[3], 10, 64)
	
	if existingList, ok := lists[listName]; ok {
		if startingIndex >= int64(len(existingList)) {
			return resp.Array([]string{})
		}
		
		if endingIndex >= int64(len(existingList)) {
			return resp.Array(existingList[startingIndex:len(existingList)])
		}
		
		if startingIndex > endingIndex {
			return resp.Array([]string{})
		}
		return resp.Array(existingList[startingIndex:endingIndex+1])
	}
	
	return resp.Array([]string{})
}
