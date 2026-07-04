package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/configs"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	flag.Parse()
	addr := fmt.Sprintf("0.0.0.0:%d", *configs.Port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to %s\n", addr)
		os.Exit(1)
	}

	if configs.MasterHost != "" {
		masterAddr := fmt.Sprintf("%s:%d", configs.MasterHost, configs.MasterPort)
		conn, err := net.Dial("tcp", masterAddr)
		if err != nil {
			os.Exit(1)
		}

		_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	}

	store := store.New()
	handler := command.NewHandler(store)

	server.Serve(l, handler)
}
