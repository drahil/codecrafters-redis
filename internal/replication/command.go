package replication

import (
	"fmt"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/configs"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func StartReplica(cfg configs.Configs) error {
	masterAddr := fmt.Sprintf("%s:%d", cfg.MasterHost, cfg.MasterPort)
	conn, err := net.Dial("tcp", masterAddr)

	if err != nil {
		return err
	}

	err = sendCommand(conn, []string{"PING"})

	if err != nil {
		return err
	}

	err = sendCommand(conn, []string{"REPLCONF", "listening-port", strconv.Itoa(cfg.Port)})

	if err != nil {
		return err
	}

	err = sendCommand(conn, []string{"REPLCONF", "capa", "psync2"})

	if err != nil {
		return err
	}

	err = sendCommand(conn, []string{"PSYNC", "?", "-1"})

	if err != nil {
		return err
	}

	return nil
}

func sendCommand(conn net.Conn, args []string) error {
	_, err := conn.Write([]byte(resp.Array(args)))
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	fmt.Println("master replied:", string(buf[:n]))
	return nil
}
