package store

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
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
	XreadWaiters   map[string][]chan struct{}
}

type BlockedClient struct {
	response chan []string
}

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

type StreamResult struct {
	Name    string
	Entries []StreamEntry
}

func New() *Store {
	return &Store{
		Strings:        make(map[string]Entry),
		Lists:          make(map[string][]string),
		BlockedClients: make(map[string][]chan []string),
		Streams:        make(map[string][]StreamEntry),
		XreadWaiters:   make(map[string][]chan struct{}),
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

func (s *Store) Get(key string) (string, bool) {
	entry, ok := s.Strings[key]
	if !ok {
		return "", false
	}

	if entry.ExpireTime != -1 && time.Now().UnixMilli() > entry.ExpireTime {
		return "", false
	}

	return entry.Value, true
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

	waiters := s.XreadWaiters[streamName]
	delete(s.XreadWaiters, streamName)

	for _, waiter := range waiters {
		select {
		case waiter <- struct{}{}:
		default:
		}
	}

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

func (s *Store) Xrange(stream, startId, endId string) []StreamEntry {
	startIndex, ok := findStreamEntry(s.Streams[stream], startId)
	if !ok {
		return []StreamEntry{}
	}

	endIndex, ok := findStreamEntry(s.Streams[stream], endId)
	if !ok {
		return []StreamEntry{}
	}

	return s.Streams[stream][startIndex : endIndex+1]
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

func (s *Store) Xread(streams, startIds []string) []StreamResult {
	return s.xreadLocked(streams, startIds)
}

func (s *Store) XreadBlock(streams, startIds []string, timeoutMs int) []StreamResult {
	waiter := make(chan struct{}, 1)

	if results := s.xreadLocked(streams, startIds); len(results) > 0 {
		return results
	}
	s.addXreadWaiter(streams, waiter)

	var timeout <-chan time.Time
	var timer *time.Timer
	if timeoutMs > 0 {
		timer = time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
		timeout = timer.C
		defer timer.Stop()
	}

	for {
		select {
		case <-waiter:
			results := s.xreadLocked(streams, startIds)
			if len(results) > 0 {
				s.removeXreadWaiter(streams, waiter)
				return results
			}

			s.removeXreadWaiter(streams, waiter)
			s.addXreadWaiter(streams, waiter)
		case <-timeout:
			s.removeXreadWaiter(streams, waiter)
			return nil
		}
	}
}

func (s *Store) addXreadWaiter(streams []string, waiter chan struct{}) {
	for _, stream := range streams {
		s.XreadWaiters[stream] = append(s.XreadWaiters[stream], waiter)
	}
}

func (s *Store) removeXreadWaiter(streams []string, waiter chan struct{}) {
	for _, stream := range streams {
		waiters := s.XreadWaiters[stream]
		for i, existing := range waiters {
			if existing == waiter {
				s.XreadWaiters[stream] = append(waiters[:i], waiters[i+1:]...)
				break
			}
		}

		if len(s.XreadWaiters[stream]) == 0 {
			delete(s.XreadWaiters, stream)
		}
	}
}

func (s *Store) xreadLocked(streams, startIds []string) []StreamResult {
	var results []StreamResult

	for i, stream := range streams {
		result, ok := s.xreadFromOneStream(stream, startIds[i])
		if ok {
			results = append(results, result)
		}
	}

	return results
}

func (s *Store) xreadFromOneStream(stream, startId string) (StreamResult, bool) {
	streamEntries := s.Streams[stream]
	startIndex, ok := findFirstStreamEntryAfter(streamEntries, startId)
	if !ok {
		return StreamResult{}, false
	}

	return StreamResult{
		Name:    stream,
		Entries: streamEntries[startIndex:],
	}, true
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
