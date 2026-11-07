package sessions

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// summaryCursorsMax bounds the cursors ListSummaries retains. A larger
// $DSH_HOME drops the oldest state and re-reads those logs from their tail
// window, so the process never grows with the number of sessions.
const summaryCursorsMax = 1024

// summaryCursor is one log's incremental read state. The lock is per path, so
// concurrent listings decode different logs in parallel and never advance the
// same cursor twice.
type summaryCursor struct {
	mu   sync.Mutex
	tail *Tail
}

var summaryCursors = struct {
	sync.Mutex
	byPath map[string]*summaryCursor
}{byPath: map[string]*summaryCursor{}}

func summaryCursorFor(path string) *summaryCursor {
	summaryCursors.Lock()
	defer summaryCursors.Unlock()
	if cursor, ok := summaryCursors.byPath[path]; ok {
		return cursor
	}
	if len(summaryCursors.byPath) >= summaryCursorsMax {
		summaryCursors.byPath = map[string]*summaryCursor{}
	}
	cursor := &summaryCursor{}
	summaryCursors.byPath[path] = cursor
	return cursor
}

// summary folds the log's header and tail window, reusing what an earlier call
// decoded. It drops the transcript rows: a listing shows no messages.
func (c *summaryCursor) summary(a artifact, size int64, mod time.Time) (Summary, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	tail := c.tail
	if tail != nil && size < tail.Size {
		// The log shrank, so its identity and events changed. Read it again.
		tail = nil
	}
	if tail == nil {
		tail = &Tail{Path: a.path, summaryOnly: true, Summary: Summary{Path: a.path, Encoding: a.encoding, FormatVersion: int(a.version), Outcome: "empty"}}
	}
	if size > tail.Offset {
		if err := tail.advance(a, size); err != nil {
			return Summary{}, err
		}
		tail.Size = size
	}
	sum := tail.Summary
	sum.SizeBytes = size
	if sum.UpdatedAt == 0 {
		sum.UpdatedAt = mod.UnixMilli()
	}
	c.tail = tail
	return sum, nil
}

// ListSummaries returns the newest-first summaries List returns, but it reads
// each log's header and tail window instead of the whole event log. An
// unchanged log costs one stat; a growing log costs only the bytes appended
// since the previous call.
//
// Outcome and Title match List unless the last turn event sits outside the tail
// window (coldTailWindow at the end of the log); UpdatedAt then falls back to
// the file modification time. Sessions in the catalogue are still listed with
// the outcome the tail shows, which is what the agents view needs.
func (s *Store) ListSummaries() ([]Summary, []string, error) {
	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return nil, nil, err
	}
	type target struct {
		a    artifact
		size int64
		mod  time.Time
	}
	targets := make([]target, 0, len(artifacts))
	for _, a := range artifacts {
		info, statErr := os.Lstat(a.path)
		if statErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", a.path, statErr))
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		targets = append(targets, target{a: a, size: info.Size(), mod: info.ModTime()})
	}
	type decoded struct {
		sum  Summary
		path string
		err  error
	}
	got := make([]decoded, len(targets))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, t := range targets {
		wg.Add(1)
		go func(i int, t target) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sum, err := summaryCursorFor(t.a.path).summary(t.a, t.size, t.mod)
			got[i] = decoded{sum: sum, path: t.a.path, err: err}
		}(i, t)
	}
	wg.Wait()
	out := make([]Summary, 0, len(got))
	ids := map[string]string{}
	for _, row := range got {
		if row.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", row.path, row.err))
			continue
		}
		if row.sum.ID == "" {
			warnings = append(warnings, fmt.Sprintf("%s: missing session header", row.path))
			continue
		}
		if prev, ok := ids[row.sum.ID]; ok {
			return nil, warnings, fmt.Errorf("duplicate session id %q in %s and %s", row.sum.ID, prev, row.path)
		}
		ids[row.sum.ID] = row.path
		out = append(out, row.sum)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt != out[j].UpdatedAt {
			return out[i].UpdatedAt > out[j].UpdatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, warnings, nil
}
