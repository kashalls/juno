// Package profile fetches and caches the tracked user's Discord user object
// (banner, accent colour, ...) - REST-only data with no gateway push
// equivalent, so it's refreshed on a timer rather than updated from events
// like presence.Store.
package profile

import "sync"

// Store caches the JSON body of the last successful profile fetch.
type Store struct {
	mu   sync.RWMutex
	last []byte
}

func NewStore() *Store {
	return &Store{}
}

// Get returns the last fetched profile JSON, or nil if none has completed yet.
func (s *Store) Get() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.last
}

func (s *Store) Set(body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = body
}
