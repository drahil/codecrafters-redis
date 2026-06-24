package command

import "github.com/codecrafters-io/redis-starter-go/internal/resp"

func (h *Handler) incr(args []string) string {
	key := args[1]
	oldValue, ok := h.store.Get(key)

	if !ok {
		oldValue = "0"
	}

	value, err := h.store.Incr(key, oldValue)
	if err != nil {
		return resp.SimpleError("value is not an integer or out of range")
	}

	return resp.Integer(value)
}
