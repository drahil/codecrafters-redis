package server

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"net"
	"os"
)

func Serve(l net.Listener, handler *command.Handler) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler *command.Handler) {
	defer conn.Close()

	for {
		args, err := resp.GetArgs(conn)

		if err != nil {
			break
		}

		if len(args) == 0 {
			continue
		}

		response := handler.Handle(args)
		conn.Write([]byte(response))
	}
}
