package command

import "github.com/codecrafters-io/redis-starter-go/internal/resp"

func (h *Handler) info(client *ClientState, args []string) string {
	if args[2] == "replication" {
		return resp.BulkString("role:master")
	}

	return ""
}