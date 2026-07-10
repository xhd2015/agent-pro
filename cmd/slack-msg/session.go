package main

import (
	"sync"
	"time"
)

type threadSession struct {
	SessionID  string
	ChannelID  string
	ThreadTS   string
	LastActive time.Time
	Started    bool
}

type sessionStore struct {
	mu          sync.Mutex
	sessions    map[string]*threadSession
	idleTimeout time.Duration
}

func newSessionStore(idleTimeout time.Duration) *sessionStore {
	return &sessionStore{
		sessions:    make(map[string]*threadSession),
		idleTimeout: idleTimeout,
	}
}

func (s *sessionStore) key(channelID, threadTS string) string {
	return channelID + ":" + threadTS
}

func (s *sessionStore) lookup(channelID, threadTS string) (*threadSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[s.key(channelID, threadTS)]
	if !ok {
		return nil, false
	}
	if s.idleTimeout > 0 && time.Since(sess.LastActive) > s.idleTimeout {
		delete(s.sessions, s.key(channelID, threadTS))
		return nil, false
	}
	return sess, true
}

func (s *sessionStore) touch(channelID, threadTS, sessionID string, started bool) *threadSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.key(channelID, threadTS)
	sess, ok := s.sessions[key]
	if !ok {
		sess = &threadSession{
			SessionID: sessionID,
			ChannelID: channelID,
			ThreadTS:  threadTS,
		}
		s.sessions[key] = sess
	}
	sess.LastActive = time.Now()
	if started {
		sess.Started = true
	}
	return sess
}

func (s *sessionStore) pruneLoop(done <-chan struct{}) {
	if s.idleTimeout <= 0 {
		return
	}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.pruneExpired()
		}
	}
}

func (s *sessionStore) pruneExpired() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, sess := range s.sessions {
		if now.Sub(sess.LastActive) > s.idleTimeout {
			delete(s.sessions, key)
		}
	}
}
