package sessions

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const inspectIndexTTL = 2 * time.Second

type inspectIndex struct {
	root string
	at   time.Time
	byID map[string]artifact
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
	inspectMemo.mu.Lock()
	idx := inspectMemo.index
	if idx != nil && idx.root == s.Root && time.Since(idx.at) < inspectIndexTTL {
		a, ok := idx.byID[id]
		inspectMemo.mu.Unlock()
		if !ok {
			return artifact{}, fmt.Errorf("DSH session %q was not found under %s", id, s.Root)
		}
		return a, nil
	}
	inspectMemo.mu.Unlock()

	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return artifact{}, err
	}
	byID := map[string]artifact{}
	var failures []string
	failures = append(failures, warnings...)
	for _, a := range artifacts {
		actual, idErr := artifactID(a)
		if idErr != nil {
			failures = append(failures, a.path+": "+idErr.Error())
			continue
		}
		if prev, ok := byID[actual]; ok {
			return artifact{}, fmt.Errorf("duplicate session id %q in %s and %s", actual, prev.path, a.path)
		}
		byID[actual] = a
	}
	inspectMemo.mu.Lock()
	inspectMemo.index = &inspectIndex{root: s.Root, at: time.Now(), byID: byID}
	inspectMemo.mu.Unlock()
	a, ok := byID[id]
	if !ok {
		detail := ""
		if len(failures) > 0 {
			detail = "; unreadable entries: " + strings.Join(failures, "; ")
		}
		return artifact{}, fmt.Errorf("DSH session %q was not found under %s%s", id, s.Root, detail)
	}
	return a, nil
}
