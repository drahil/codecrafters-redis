package store

type Entry struct {
	Value      string
	ExpireTime int64
}

type Store struct {
	Strings map[string]Entry
	Lists   map[string][]string
}

func New() *Store {
	return &Store{
		Strings: make(map[string]Entry),
		Lists:   make(map[string][]string),
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
	s.Lists[key] = append(s.Lists[key], values...)
	return len(s.Lists[key])
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

	if start > end {
		return []string{}
	}

	return list[start : end+1]
}
