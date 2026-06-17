package command

import "strconv"

func (h *Handler) incr(args []string) int {
	key := args[1]
	value, ok := h.store.Get(key)

	if !ok {
		h.store.Set(key, "1", -1)

		return 1
	}

	oldValueInt, _ := strconv.Atoi(value)

	newValue := oldValueInt + 1

	h.store.Set(args[1], strconv.Itoa(newValue), -1)

	return newValue
}