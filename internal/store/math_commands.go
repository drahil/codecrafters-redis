package store

import "strconv"

func (s *Store) Incr(key, oldValue string) int {
	oldValueInt, _ := strconv.Atoi(oldValue)

	newValue := oldValueInt + 1

	s.Strings[key] = // not sure how to do this

	return newValue
}