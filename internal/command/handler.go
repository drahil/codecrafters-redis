package command

import (
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/replication"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Handler struct {
	store              *store.Store
	role               string
	replicationManager *replication.Manager
}

func NewHandler(store *store.Store, role string, replicationManager *replication.Manager) *Handler {
	return &Handler{
		store:              store,
		role:               role,
		replicationManager: replicationManager,
	}
}

func (h *Handler) Handle(args []string, client *ClientState, conn net.Conn) string {
	command := strings.ToLower(args[0])

	if h.CheckIfQueueIsActive(client, command, args) {
		return resp.SimpleString("QUEUED")
	}

	switch command {
	case "ping":
		return resp.SimpleString("PONG")
	case "echo":
		return resp.BulkString(args[1])
	case "set":
		response := h.set(args)
		h.propagateWriteCommand(client, args)
		return response
	case "get":
		return h.get(args[1])
	case "rpush":
		return h.rpush(args)
	case "lrange":
		return h.lrange(args)
	case "lpush":
		return h.lpush(args)
	case "llen":
		return h.llen(args)
	case "lpop":
		return h.lpop(args)
	case "blpop":
		return h.blpop(args)
	case "type":
		return h.getType(args)
	case "xadd":
		return h.xadd(args)
	case "xrange":
		return h.xrange(args)
	case "xread":
		return h.xread(args)
	case "incr":
		return h.incr(args)
	case "multi":
		return h.multi(client, args)
	case "exec":
		return h.exec(client, args)
	case "discard":
		return h.discard(client, args)
	case "info":
		return h.info(client, args)
	case "replconf":
		return h.replconf(client, args)
	case "psync":
		return h.psync(client, conn, args)

	}

	return resp.SimpleString("OK")
}

func (h *Handler) propagateWriteCommand(client *ClientState, args []string) {
	if h.role != "master" || client.IsReplica {
		return
	}

	h.replicationManager.Propagate(args)
}

func (h *Handler) CheckIfQueueIsActive(client *ClientState, command string, args []string) bool {
	if client.InMulti && command != "exec" && command != "multi" && command != "discard" {
		client.Queue = append(client.Queue, args)

		return true
	}

	return false
}
