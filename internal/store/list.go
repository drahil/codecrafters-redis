package store

import "math"

type BlockedClient struct {
	response chan []string
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
