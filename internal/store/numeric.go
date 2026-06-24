package store

import "strconv"

func (s *Store) Incr(key, oldValue string) (int, error) {
	oldValueInt, err := strconv.Atoi(oldValue)
	if err != nil {
		return 0, err
	}

	newValue := oldValueInt + 1

	s.Set(key, strconv.Itoa(newValue), -1)

	return newValue, nil
}
