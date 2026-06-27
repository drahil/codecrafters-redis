package command

import "github.com/codecrafters-io/redis-starter-go/internal/resp"

func (h *Handler) multi(client *ClientState, args []string) string {
	client.InitializeMulti()

	return resp.SimpleString("OK")
}

func (h *Handler) exec(client *ClientState, args []string) string {
	if !client.InMulti {
		return resp.SimpleError("EXEC without MULTI")
	}

	queuedCommands := client.Queue
	client.InMulti = false
	client.Queue = nil

	results := make([]string, 0, len(queuedCommands))
	for _, queuedCommand := range queuedCommands {
		result := h.Handle(queuedCommand, client)
		results = append(results, result)
	}

	return resp.RawArray(results)
}

func (h *Handler) discard(client *ClientState, args []string) string {
	if !client.InMulti {
		return resp.SimpleError("DISCARD without MULTI")
	}

	client.InMulti = false
	client.Queue = nil

	return resp.SimpleString("OK")
}
