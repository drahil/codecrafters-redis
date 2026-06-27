package command

import (
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Handle(args []string, client *ClientState) string {
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
		return h.set(args)
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
	}

	return resp.SimpleString("OK")
}

func (h *Handler) CheckIfQueueIsActive(client *ClientState, command string, args []string) bool {
	if client.InMulti && command != "exec" && command != "multi" {
		client.Queue = append(client.Queue, args)

		return true
	}

	return false
}
