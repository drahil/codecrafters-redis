package command

import (
	"strconv"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (h *Handler) incr(args []string) int {
	key := args[1]
	oldValue, ok := h.store.Get(key)

	if !ok {
		oldValue = "0"
	}
	
	value := h.store.Incr(key, oldValue)
	
	return resp.Integer(value)
}