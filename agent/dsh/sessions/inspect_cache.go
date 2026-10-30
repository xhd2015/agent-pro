package sessions

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// inspectIndexTTL is longer than the orchestrator poll (3s) so a warm server
// does not re-walk every session log on each refresh. A lookup that misses
// still rebuilds, so a session created after the index was cached is visible
// on the next read.
const inspectIndexTTL = 15 * time.Second

type inspectIndex struct {
	root     string
	at       time.Time
	byID     map[string]artifact
	failures []string
}

var inspectMemo = struct {
	mu    sync.Mutex
	index *inspectIndex
}{}

// Exists is a cheap id lookup (artifact index only, not Inspect).
func (s *Store) Exists(id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}
	_, err := s.lookupArtifact(id)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "was not found") {
		return false, nil
	}
	return false, err
}

func (s *Store) lookupArtifact(id string) (artifact, error) {
	// Hold the lock across the rebuild so concurrent probes share one walk
	// instead of each re-reading every session header.
	inspectMemo.mu.Lock()
	defer inspectMemo.mu.Unlock()
	if a, ok := s.cachedArtifact(id); ok {
		return a, nil
	}
	idx, err := s.buildArtifactIndex()
	if err != nil {
		return artifact{}, err
	}
	inspectMemo.index = idx
	a, ok := idx.byID[id]
	if !ok {
		detail := ""
		if len(idx.failures) > 0 {
			detail = "; unreadable entries: " + strings.Join(idx.failures, "; ")
		}
		return artifact{}, fmt.Errorf("DSH session %q was not found under %s%s", id, s.Root, detail)
	}
	return a, nil
}

func (s *Store) cachedArtifact(id string) (artifact, bool) {
	idx := inspectMemo.index
	if idx == nil || idx.root != s.Root || time.Since(idx.at) >= inspectIndexTTL {
		return artifact{}, false
	}
	a, ok := idx.byID[id]
	return a, ok
}

func (s *Store) buildArtifactIndex() (*inspectIndex, error) {
	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return nil, err
	}
	byID := map[string]artifact{}
	failures := append([]string(nil), warnings...)
	for _, a := range artifacts {
		actual, idErr := artifactID(a)
		if idErr != nil {
			failures = append(failures, a.path+": "+idErr.Error())
			continue
		}
		if prev, ok := byID[actual]; ok {
			return nil, fmt.Errorf("duplicate session id %q in %s and %s", actual, prev.path, a.path)
		}
		byID[actual] = a
	}
	return &inspectIndex{root: s.Root, at: time.Now(), byID: byID, failures: failures}, nil
}
