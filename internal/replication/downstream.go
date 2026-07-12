package replication

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

const MasterReplID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
const emptyRDBHex = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

var emptyRDB, _ = hex.DecodeString(emptyRDBHex)

// DownstreamReplicas tracks replicas connected to this server when it is master.
type DownstreamReplicas struct {
	replicas []net.Conn
}

func NewDownstreamReplicas() *DownstreamReplicas {
	return &DownstreamReplicas{}
}

func (d *DownstreamReplicas) AddReplica(conn net.Conn) {
	d.replicas = append(d.replicas, conn)
}

func (d *DownstreamReplicas) StartFullSync(conn net.Conn) string {
	d.AddReplica(conn)

	fullResync := resp.SimpleString("FULLRESYNC " + MasterReplID + " 0")
	rdbFile := fmt.Sprintf("$%d\r\n%s", len(emptyRDB), string(emptyRDB))

	return fullResync + rdbFile
}

func (d *DownstreamReplicas) PropagateCommand(args []string) {
	encoded := []byte(resp.Array(args))

	for _, replica := range d.replicas {
		replica.Write(encoded)
	}
}

func MasterInfo() string {
	return "role:master\nmaster_replid:" + MasterReplID + "\nmaster_repl_offset:0"
}
