package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (h *Handler) set(args []string) string {
	var expireTime int64 = -1

	if len(args) > 3 && strings.ToLower(args[3]) == "ex" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		expireTime *= 1000
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}
	if len(args) > 4 && strings.ToLower(args[3]) == "px" {
		expireTime, _ = strconv.ParseInt(args[4], 10, 64)
		nowMs := time.Now().UnixMilli()
		expireTime = expireTime + nowMs
	}

	h.store.Set(args[1], args[2], expireTime)

	return resp.SimpleString("OK")
}

func (h *Handler) get(key string) string {
	value, ok := h.store.Get(key)

	if !ok {
		return resp.NullBulkString()
	}

	return resp.BulkString(value)
}

func (h *Handler) getType(args []string) string {
	return resp.SimpleString(h.store.Type(args[1]))
}
