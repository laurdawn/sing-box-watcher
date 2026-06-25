package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const sessionTTL = 24 * time.Hour

type Store struct {
	mu       sync.RWMutex
	sessions map[string]time.Time
}

func NewStore() *Store {
	s := &Store{sessions: make(map[string]time.Time)}
	go s.cleanLoop()
	return s
}

func (s *Store) Create() string {
	token := randomHex(32)
	s.mu.Lock()
	s.sessions[token] = time.Now().Add(sessionTTL)
	s.mu.Unlock()
	return token
}

func (s *Store) Valid(token string) bool {
	s.mu.RLock()
	exp, ok := s.sessions[token]
	s.mu.RUnlock()
	return ok && time.Now().Before(exp)
}

func (s *Store) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *Store) cleanLoop() {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		s.mu.Lock()
		for tok, exp := range s.sessions {
			if now.After(exp) {
				delete(s.sessions, tok)
			}
		}
		s.mu.Unlock()
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
