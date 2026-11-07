package sessions

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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

type artifactIDMemo struct {
	id   string
	size int64
}

var artifactIDs = struct {
	mu     sync.Mutex
	byPath map[string]artifactIDMemo
}{byPath: map[string]artifactIDMemo{}}

// ArtifactIDReads counts header decodes (zstd/jsonl first line). Tests use it
// to prove an appended log does not re-read the session id.
var ArtifactIDReads atomic.Int64

func resetArtifactIDs() {
	artifactIDs.mu.Lock()
	artifactIDs.byPath = map[string]artifactIDMemo{}
	artifactIDs.mu.Unlock()
	ArtifactIDReads.Store(0)
}

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
	seen := map[string]struct{}{}
	for _, a := range artifacts {
		seen[a.path] = struct{}{}
		actual, idErr := artifactIDCached(a)
		if idErr != nil {
			failures = append(failures, a.path+": "+idErr.Error())
			continue
		}
		if prev, ok := byID[actual]; ok {
			return nil, fmt.Errorf("duplicate session id %q in %s and %s", actual, prev.path, a.path)
		}
		byID[actual] = a
	}
	artifactIDs.mu.Lock()
	for path := range artifactIDs.byPath {
		if _, ok := seen[path]; !ok {
			delete(artifactIDs.byPath, path)
		}
	}
	artifactIDs.mu.Unlock()
	return &inspectIndex{root: s.Root, at: time.Now(), byID: byID, failures: failures}, nil
}

// artifactIDCached reuses the header id while a log only grows. Session id is
// in the first line and does not change on append; a shrink or replace re-reads.
func artifactIDCached(a artifact) (string, error) {
	info, err := os.Lstat(a.path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("selected artifact is not a regular file")
	}
	size := info.Size()
	artifactIDs.mu.Lock()
	if memo, ok := artifactIDs.byPath[a.path]; ok && size >= memo.size {
		if size > memo.size {
			memo.size = size
			artifactIDs.byPath[a.path] = memo
		}
		artifactIDs.mu.Unlock()
		return memo.id, nil
	}
	artifactIDs.mu.Unlock()
	ArtifactIDReads.Add(1)
	id, err := artifactID(a)
	if err != nil {
		return "", err
	}
	if st, statErr := os.Lstat(a.path); statErr == nil {
		size = st.Size()
	}
	artifactIDs.mu.Lock()
	artifactIDs.byPath[a.path] = artifactIDMemo{id: id, size: size}
	artifactIDs.mu.Unlock()
	return id, nil
}
