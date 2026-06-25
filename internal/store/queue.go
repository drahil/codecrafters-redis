package store

func (s *Store) InitializeMulti() {
	s.MultiInitialized = true
	s.QueuedCommands = nil
}
