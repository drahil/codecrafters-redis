package main

import (
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/configs"
	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	cfg := configs.Init()

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to %s\n", addr)
		os.Exit(1)
	}

	if cfg.MasterHost != "" {
		if err := replication.StartReplica(cfg); err != nil {
			os.Exit(1)
		}
	}

	role := "master"
	if cfg.ReplicaOf != "" {
		role = "replica"
	}

	store := store.New()
	replicationManager := replication.NewManager()

	handler := command.NewHandler(store, role, replicationManager)

	server.Serve(l, handler)
}
