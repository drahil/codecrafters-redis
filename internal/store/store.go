package store

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Entry struct {
	Value      string
	ExpireTime int64
}

type Store struct {
	Strings        map[string]Entry
	Lists          map[string][]string
	BlockedClients map[string][]chan []string
	Streams        map[string][]StreamEntry
}

type BlockedClient struct {
	response chan []string
}

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

func New() *Store {
	return &Store{
		Strings:        make(map[string]Entry),
		Lists:          make(map[string][]string),
		BlockedClients: make(map[string][]chan []string),
		Streams:        make(map[string][]StreamEntry),
	}
}

func (s *Store) AddBlockedClient(listName string, ch chan []string) {
	s.BlockedClients[listName] = append(s.BlockedClients[listName], ch)
}

func (s *Store) RemoveBlockedClient(listName string, ch chan []string) {
	clients := s.BlockedClients[listName]
	for i, client := range clients {
		if client == ch {
			s.BlockedClients[listName] = append(clients[:i], clients[i+1:]...)
			return
		}
	}
}

func (s *Store) Set(key, value string, expireTime int64) {
	s.Strings[key] = Entry{
		Value:      value,
		ExpireTime: expireTime,
	}
}

func (s *Store) Get(key string) (Entry, bool) {
	entry, ok := s.Strings[key]
	return entry, ok
}

func (s *Store) RPush(key string, values ...string) int {
	consumed := 0

	for _, value := range values {
		if len(s.BlockedClients[key]) > 0 {
			ch := s.BlockedClients[key][0]
			s.BlockedClients[key] = s.BlockedClients[key][1:]

			ch <- []string{key, value}
			consumed++
			continue
		}

		s.Lists[key] = append(s.Lists[key], value)
	}

	return len(s.Lists[key]) + consumed
}

func (s *Store) LPush(key string, values ...string) int {
	consumed := 0

	for _, value := range values {
		if len(s.BlockedClients[key]) > 0 {
			ch := s.BlockedClients[key][0]
			s.BlockedClients[key] = s.BlockedClients[key][1:]
			ch <- []string{key, value}
			consumed++
			continue
		}

		s.Lists[key] = append([]string{value}, s.Lists[key]...)
	}

	return len(s.Lists[key]) + consumed
}

func (s *Store) LRange(key string, start, end int) []string {
	list, ok := s.Lists[key]
	if !ok {
		return []string{}
	}

	if start >= len(list) {
		return []string{}
	}

	if end >= len(list) {
		end = len(list) - 1
	}

	if start > end && start > 0 && end > 0 {
		return []string{}
	}

	if start < 0 && int(math.Abs(float64(start))) > len(list) {
		start = 0
	}

	if end < 0 && int(math.Abs(float64(end))) > len(list) {
		end = 0
	}

	if start < 0 {
		start = len(list) + start
	}

	if end < 0 {
		end = len(list) + end
	}

	return list[start : end+1]
}

func (s *Store) Length(listName string) int {
	return len(s.Lists[listName])
}

func (s *Store) LPop(listName string, count int) []string {
	result := s.Lists[listName][:count]
	s.Lists[listName] = s.Lists[listName][count:]

	return result
}

func (s *Store) Type(key string) string {
	if _, ok := s.Strings[key]; ok {
		return "string"
	}

	if _, ok := s.Lists[key]; ok {
		return "list"
	}

	if _, ok := s.Streams[key]; ok {
		return "stream"
	}

	return "none"
}

func (s *Store) NewStream(streamName, id, key, value string) string {
	s.Streams[streamName] = append(s.Streams[streamName], StreamEntry{
		ID: id,
		Fields: map[string]string{
			key: value,
		},
	})

	return id
}

func (s *Store) ValidateIdForStream(streamName, id string) (string, error) {
	if id == "0-0" {
		err := errors.New("The ID specified in XADD must be greater than 0-0")
		return id, err
	}

	idMicroSeconds, idSequenceNumber := parseStreamID(id)

	if idMicroSeconds == "*" {
		idMicroSeconds = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	if len(s.Streams[streamName]) == 0 {
		if idSequenceNumber == "*" {
			idSequenceNumber = "0"

			if idMicroSeconds == "0" {
				idSequenceNumber = "1"
			}
		}
		id = fmt.Sprintf("%s-%s", idMicroSeconds, idSequenceNumber)
		return id, nil
	}

	latestId := s.Streams[streamName][len(s.Streams[streamName])-1].ID
	latestIdMicroSeconds, latestIdSequenceNumber := parseStreamID(latestId)

	if idSequenceNumber == "*" {
		idSequenceNumber = "0"

		if latestIdMicroSeconds == idMicroSeconds {
			idSequenceNumber = "1"
		}

		if latestIdSequenceNumber == "0" {
			idSequenceNumber = "1"
		}
	}

	latestIdMicroSecondsInt, _ := strconv.Atoi(latestIdMicroSeconds)
	idMicroSecondsInt, _ := strconv.Atoi(idMicroSeconds)
	latestIdSequenceNumberInt, _ := strconv.Atoi(latestIdSequenceNumber)
	idSequenceNumberInt, _ := strconv.Atoi(idSequenceNumber)

	if latestIdMicroSecondsInt > idMicroSecondsInt {
		err := errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		return id, err
	}

	if latestIdMicroSecondsInt == idMicroSecondsInt && latestIdSequenceNumberInt >= idSequenceNumberInt {
		err := errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		return id, err
	}

	id = fmt.Sprintf("%s-%s", idMicroSeconds, idSequenceNumber)

	return id, nil
}

func parseStreamID(id string) (string, string) {
	parts := strings.SplitN(id, "-", 2)

	if len(parts) != 2 {
		return "*", "*"
	}

	milliseconds := parts[0]
	sequenceNumber := parts[1]

	return milliseconds, sequenceNumber
}

func (s *Store) Xrange(stream, startId, endId string) string {
	startIndex, _ := findStreamEntry(s.Streams[stream], startId)
	endIndex, _ := findStreamEntry(s.Streams[stream], endId)

	entries := s.Streams[stream][startIndex : endIndex+1]
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

func findStreamEntry(entries []StreamEntry, id string) (int, bool) {
	if id == "-" {
		return 0, true
	}

	if id == "+" {
		return len(entries) - 1, true
	}

	for i := range entries {
		if entries[i].ID == id {
			return i, true
		}
	}

	return -1, false
}

func (s *Store) Xread(stream, startId string) string {
	streamEntries := s.Streams[stream]
	startIndex, ok := findFirstStreamEntryAfter(streamEntries, startId)

	if !ok {
		return resp.NullArray()
	}

	entries := streamEntries[startIndex:]
	var builder strings.Builder

	builder.WriteString("*1\r\n")
	builder.WriteString("*2\r\n")
	builder.WriteString(resp.BulkString(stream))
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

func findFirstStreamEntryAfter(entries []StreamEntry, id string) (int, bool) {
	for i := range entries {
		if compareStreamIDs(entries[i].ID, id) > 0 {
			return i, true
		}
	}

	return -1, false
}

func compareStreamIDs(a, b string) int {
	aMs, aSeq := parseStreamID(a)
	bMs, bSeq := parseStreamID(b)

	aMsInt, _ := strconv.Atoi(aMs)
	bMsInt, _ := strconv.Atoi(bMs)

	if aMsInt < bMsInt {
		return -1
	}
	if aMsInt > bMsInt {
		return 1
	}

	aSeqInt, _ := strconv.Atoi(aSeq)
	bSeqInt, _ := strconv.Atoi(bSeq)

	if aSeqInt < bSeqInt {
		return -1
	}
	if aSeqInt > bSeqInt {
		return 1
	}

	return 0
}
