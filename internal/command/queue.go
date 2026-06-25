package command

import "github.com/codecrafters-io/redis-starter-go/internal/resp"

func (h *Handler) multi(args []string) string {
	h.store.InitializeMulti()

	return resp.SimpleString("OK")
}

func (h *Handler) exec(args []string) string {
	if !h.store.MultiInitialized {
		return resp.SimpleError("EXEC without MULTI")
	}

	queuedCommands := h.store.QueuedCommands
	h.store.MultiInitialized = false
	h.store.QueuedCommands = nil

	results := make([]string, 0, len(queuedCommands))
	for _, queuedCommand := range queuedCommands {
		result := h.Handle(queuedCommand)
		results = append(results, result)
	}

	return resp.RawArray(results)
}
