package store

import "time"

type Entry struct {
	Value      string
	ExpireTime int64
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
