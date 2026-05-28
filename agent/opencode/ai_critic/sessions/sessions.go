package sessions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type SessionData struct {
	ID         string `json:"id"`
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	ProjectDir string `json:"project_dir"`
	Port       int    `json:"port"`
	CreatedAt  string `json:"created_at"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

func Load(dir, sessionID string) (*SessionData, error) {
	path := filepath.Join(dir, sessionID, "session.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s SessionData
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func Save(dir string, s *SessionData) error {
	sessionDir := filepath.Join(dir, s.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionDir, "session.json"), data, 0644)
}

func List(dir string) ([]SessionData, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var sessions []SessionData
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		s, err := Load(dir, entry.Name())
		if err != nil || s == nil {
			continue
		}
		sessions = append(sessions, *s)
	}

	slices.SortFunc(sessions, func(a, b SessionData) int {
		ta, _ := time.Parse(time.RFC3339, a.CreatedAt)
		tb, _ := time.Parse(time.RFC3339, b.CreatedAt)
		if tb.Before(ta) {
			return -1
		}
		if ta.Before(tb) {
			return 1
		}
		return 0
	})
	return sessions, nil
}

func UpdateStatus(dir, sessionID, status, errMsg string) error {
	s, err := Load(dir, sessionID)
	if err != nil || s == nil {
		return err
	}
	s.Status = status
	s.Error = errMsg
	return Save(dir, s)
}

func Delete(dir, sessionID string) error {
	return os.RemoveAll(filepath.Join(dir, sessionID))
}

type Store[T any] struct {
	mu   sync.RWMutex
	path string
}

func NewStore[T any](path string) *Store[T] {
	return &Store[T]{path: path}
}

func (s *Store[T]) Load() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []T
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store[T]) Save(entries []T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store[T]) Add(entry T) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.loadLocked()
	if err != nil {
		return err
	}
	entries = append(entries, entry)
	return s.saveLocked(entries)
}

func (s *Store[T]) Update(fn func(entries *[]T) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.loadLocked()
	if err != nil {
		return err
	}
	if err := fn(&entries); err != nil {
		return err
	}
	return s.saveLocked(entries)
}

func (s *Store[T]) loadLocked() ([]T, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []T
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store[T]) saveLocked(entries []T) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
