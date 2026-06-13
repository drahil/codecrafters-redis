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
		return resp.NullBulkString()
	}

	if entry.ExpireTime == -1 {
		return resp.BulkString(entry.Value)
	}

	nowMs := time.Now().UnixMilli()

	if nowMs <= entry.ExpireTime {
		return resp.BulkString(entry.Value)
	}

	return resp.NullBulkString()
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

func (h *Handler) lpush(args []string) string {
	listName := args[1]
	values := args[2:]

	length := h.store.LPush(listName, values...)

	return resp.Integer(length)
}

func (h *Handler) llen(args []string) string {
	listName := args[1]

	return resp.Integer(h.store.Length(listName))
}

func (h *Handler) lpop(args []string) string {
	listName := args[1]
	numberOfElementsToRemove := 1

	if len(args) == 3 {
		numberOfElementsToRemove, _ = strconv.Atoi(args[2])
	}

	if listName == "" {
		return resp.NullBulkString()
	}

	if h.store.Length(listName) == 0 {
		return resp.NullBulkString()
	}

	if h.store.Length(listName) < numberOfElementsToRemove {
		numberOfElementsToRemove = h.store.Length(listName)
	}

	result := h.store.LPop(listName, numberOfElementsToRemove)

	if len(result) == 1 {
		return resp.BulkString(result[0])
	}
	return resp.Array(result)
}

func (h *Handler) blpop(args []string) string {
	listName := args[1]

	if h.store.Length(listName) > 0 {
		result := h.store.LPop(listName, 1)
		return resp.Array([]string{listName, result[0]})
	}

	timeoutSeconds, _ := strconv.ParseFloat(args[2], 64)

	ch := make(chan []string)
	h.store.AddBlockedClient(listName, ch)

	if timeoutSeconds == 0 {
		result := <-ch
		return resp.Array(result)
	}

	select {
	case result := <-ch:
		return resp.Array(result)
	case <-time.After(time.Duration(timeoutSeconds * float64(time.Second))):
		h.store.RemoveBlockedClient(listName, ch)
		return resp.NullArray()
	}
}

func (h *Handler) getType(args []string) string {
	return resp.SimpleString(h.store.Type(args[1]))
}

func (h *Handler) xadd(args []string) string {
	stream := args[1]
	id := args[2]
	key := args[3]
	value := args[4]

	id, err := h.store.ValidateIdForStream(stream, id)

	if err != nil {
		return resp.SimpleError(err.Error())
	}

	response := h.store.NewStream(stream, id, key, value)
	return resp.BulkString(response)
}

func (h *Handler) xrange(args []string) string {
	stream := args[1]
	startId := args[2]
	endId := args[3]

	return h.store.Xrange(stream, startId, endId)
}

func (h *Handler) xread(args []string) string {
	if len(args) > 3 && args[1] == "block" {
		timeoutMs, _ := strconv.Atoi(args[2])
		streamCount := (len(args) - 4) / 2
		streams := args[4 : 4+streamCount]
		startIds := args[4+streamCount:]

		return h.store.XreadBlock(streams, startIds, timeoutMs)
	}

	streamCount := (len(args) - 2) / 2
	streams := args[2 : 2+streamCount]
	startIds := args[2+streamCount:]

	return h.store.Xread(streams, startIds)
}
