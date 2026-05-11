package job

func (s *Store) Subscribe(id string) (<-chan Response, func(), bool) {
	s.mu.Lock()
	if _, ok := s.jobs[id]; !ok {
		s.mu.Unlock()
		return nil, nil, false
	}
	ch := make(chan Response, 16)
	if s.listeners[id] == nil {
		s.listeners[id] = map[chan Response]struct{}{}
	}
	s.listeners[id][ch] = struct{}{}
	initial := s.responseLocked(id)
	s.mu.Unlock()
	ch <- initial
	return ch, func() { s.unsubscribe(id, ch) }, true
}

func (s *Store) unsubscribe(id string, ch chan Response) {
	s.mu.Lock()
	if listeners := s.listeners[id]; listeners != nil {
		delete(listeners, ch)
		if len(listeners) == 0 {
			delete(s.listeners, id)
		}
	}
	close(ch)
	s.mu.Unlock()
}

func (s *Store) update(id string, mutate func(*Record)) bool {
	var snapshot *Record
	s.mu.Lock()
	record, ok := s.jobs[id]
	if ok {
		mutate(record)
		snapshot = cloneRecord(record)
	}
	s.mu.Unlock()
	if ok {
		s.broadcast(id)
		s.notify(snapshot)
	}
	return ok
}

func (s *Store) broadcast(id string) {
	s.mu.RLock()
	response := s.responseLocked(id)
	for ch := range s.listeners[id] {
		select {
		case ch <- response:
		default:
		}
	}
	s.mu.RUnlock()
}

func (s *Store) notify(record *Record) {
	if record == nil {
		return
	}
	for _, sink := range s.sinks {
		sink.SaveJob(cloneRecord(record))
	}
}

func (s *Store) responseLocked(id string) Response {
	record := s.jobs[id]
	if record == nil {
		return Response{ID: id, Status: StatusFailed}
	}
	return s.toResponse(record)
}
