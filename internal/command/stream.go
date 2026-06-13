package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

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

	return encodeStreamEntries(h.store.Xrange(stream, startId, endId))
}

func (h *Handler) xread(args []string) string {
	if len(args) > 3 && strings.ToLower(args[1]) == "block" {
		timeoutMs, _ := strconv.Atoi(args[2])
		streamCount := (len(args) - 4) / 2
		streams := args[4 : 4+streamCount]
		startIds := args[4+streamCount:]

		return encodeStreamResults(h.store.XreadBlock(streams, startIds, timeoutMs))
	}

	streamCount := (len(args) - 2) / 2
	streams := args[2 : 2+streamCount]
	startIds := args[2+streamCount:]

	return encodeStreamResults(h.store.Xread(streams, startIds))
}

func encodeStreamResults(results []store.StreamResult) string {
	if len(results) == 0 {
		return resp.NullArray()
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(results)))

	for _, result := range results {
		builder.WriteString("*2\r\n")
		builder.WriteString(resp.BulkString(result.Name))
		builder.WriteString(encodeStreamEntries(result.Entries))
	}

	return builder.String()
}

func encodeStreamEntries(entries []store.StreamEntry) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))

	for _, entry := range entries {
		builder.WriteString("*2\r\n")
		builder.WriteString(resp.BulkString(entry.ID))
		builder.WriteString(fmt.Sprintf("*%d\r\n", len(entry.Fields)*2))

		for key, value := range entry.Fields {
			builder.WriteString(resp.BulkString(key))
			builder.WriteString(resp.BulkString(value))
		}
	}

	return builder.String()
}
