package command

import (
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

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
