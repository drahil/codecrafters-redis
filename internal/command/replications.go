package command

import (
	"encoding/hex"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

const masterReplID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
const emptyRDBHex = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

var emptyRDB, _ = hex.DecodeString(emptyRDBHex)

func (h *Handler) info(client *ClientState, args []string) string {
	if args[1] == "replication" {
		if h.role != "master" {
			return resp.BulkString("role:slave")
		}

		return resp.BulkString("role:master\nmaster_replid:" + masterReplID + "\nmaster_repl_offset:0")
	}

	return ""
}

func (h *Handler) replconf(client *ClientState, args []string) string {
	return resp.SimpleString("OK")
}

func (h *Handler) psync(client *ClientState, args []string) string {
	fullResync := resp.SimpleString("FULLRESYNC " + masterReplID + " 0")
	rdbFile := fmt.Sprintf("$%d\r\n%s", len(emptyRDB), string(emptyRDB))

	return fullResync + rdbFile
}
