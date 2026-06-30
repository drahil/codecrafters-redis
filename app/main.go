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
	
	addr := fmt.Sprintf("0.0.0.0:%d", *configs.Port)
	flag.Parse()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	store := store.New()
	handler := command.NewHandler(store)

	server.Serve(l, handler)
}
