package command

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Handle(args []string) string {
	switch args[0] {
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
	}

	return resp.SimpleString("OK")
}

func (h *Handler) set(args []string) string {
	var expireTime int64 = -1

	if len(args) > 3 && args[3] == "ex" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		expireTime *= 1000
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}
	if len(args) > 4 && args[3] == "px" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}

	h.store.Set(args[1], args[2], expireTime)

	return resp.SimpleString("OK")
}

func (h *Handler) get(key string) string {
	entry, _ := h.store.Get(key)

	if entry.Value == "" {
		return "$-1\r\n"
	}

	if entry.ExpireTime == -1 {
		return resp.BulkString(entry.Value)
	}

	nowMs := time.Now().UnixMilli()

	if nowMs <= entry.ExpireTime {
		return resp.BulkString(entry.Value)
	}

	return "$-1\r\n"
}

func (h *Handler) rpush(args []string) string {
	listName := args[1]
	values := args[2:]

	length := h.store.RPush(listName, values...)

	return resp.Integer(length)
}

func (h *Handler) lrange(args []string) string {
	listName := args[1]

	start, _ := strconv.Atoi(args[2])
	end, _ := strconv.Atoi(args[3])

	values := h.store.LRange(listName, start, end)

	return resp.Array(values)
}
