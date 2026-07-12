package command

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (h *Handler) info(client *ClientState, args []string) string {
	if args[1] == "replication" {
		if h.role != "master" {
			return resp.BulkString("role:slave")
		}

		return resp.BulkString(replication.MasterInfo())
	}

	return ""
}

func (h *Handler) replconf(client *ClientState, args []string) string {
	return resp.SimpleString("OK")
}

func (h *Handler) psync(client *ClientState, conn net.Conn, args []string) string {
	client.IsReplica = true

	return h.downstreamReplicas.StartFullSync(conn)
}
