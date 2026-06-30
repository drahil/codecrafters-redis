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
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	flag.Parse()
	addr := fmt.Sprintf("0.0.0.0:%d", *configs.Port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to %s\n", addr)
		os.Exit(1)
	}

	store := store.New()
	handler := command.NewHandler(store)

	server.Serve(l, handler)
}
