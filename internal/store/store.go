package store

type Store struct {
	Strings        map[string]Entry
	Lists          map[string][]string
	BlockedClients map[string][]chan []string
	Streams        map[string][]StreamEntry
	XreadWaiters   map[string][]chan struct{}
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
