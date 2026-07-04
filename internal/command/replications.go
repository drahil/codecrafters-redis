package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/configs"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (h *Handler) info(client *ClientState, args []string) string {
	if args[1] == "replication" {
		if *configs.ReplicaOf != "" {
			return resp.BulkString("role:slave")
		}
		
		return resp.BulkString("role:master\nmaster_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb\nmaster_repl_offset:0")
	}

	return ""
}