package replication

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/configs"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func StartReplica(cfg configs.Configs, processCommand func([]string)) error {
	masterAddr := fmt.Sprintf("%s:%d", cfg.MasterHost, cfg.MasterPort)
	conn, err := net.Dial("tcp", masterAddr)

	if err != nil {
		return err
	}

	reader := resp.NewReader(conn)

	err = sendCommand(conn, reader, []string{"PING"})

	if err != nil {
		return err
	}

	err = sendCommand(conn, reader, []string{"REPLCONF", "listening-port", strconv.Itoa(cfg.Port)})

	if err != nil {
		return err
	}

	err = sendCommand(conn, reader, []string{"REPLCONF", "capa", "psync2"})

	if err != nil {
		return err
	}

	err = sendPsync(conn, reader)

	if err != nil {
		return err
	}

	for {
		args, err := reader.ReadArray()
		if err != nil {
			return err
		}

		processCommand(args)
	}
}

func sendCommand(conn net.Conn, reader *resp.Reader, args []string) error {
	_, err := conn.Write([]byte(resp.Array(args)))
	if err != nil {
		return err
	}

	reply, err := reader.ReadLine()
	if err != nil {
		return err
	}

	fmt.Println("master replied:", reply)
	return nil
}

func sendPsync(conn net.Conn, reader *resp.Reader) error {
	_, err := conn.Write([]byte(resp.Array([]string{"PSYNC", "?", "-1"})))
	if err != nil {
		return err
	}

	reply, err := reader.ReadLine()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(reply, "+FULLRESYNC") {
		return fmt.Errorf("expected FULLRESYNC, got %q", reply)
	}

	fmt.Println("master replied:", reply)

	if _, err := reader.ReadBulkString(); err != nil {
		return err
	}

	return nil
}
