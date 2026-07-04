package replication

import (
	"fmt"
	"net"
)

func StartReplica(masterHost string, masterPort int) error {
	masterAddr := fmt.Sprintf("%s:%d", masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))

	if err != nil {
		return err
	}

	return nil
}
