package sessions

import (
	"fmt"
	"os"
	"sync"
)

// headerCache memoizes decoded header rows. The header is the first row of an
// append-only artifact, so growth never invalidates it; only a smaller file or
// a different identity means the artifact was replaced.
var headerCache = struct {
	sync.Mutex
	byPath map[string]headerEntry
}{byPath: map[string]headerEntry{}}

type headerEntry struct {
	size    int64
	summary Summary
}

func resetHeaderCache() {
	headerCache.Lock()
	headerCache.byPath = map[string]headerEntry{}
	headerCache.Unlock()
}

// Header returns a session's header summary — identity, cwd, lineage and
// format — without decoding the event log. File previews, directory scopes and
// skill lookups need only these fields; on a live session that is still
// appending, Inspect would re-decode the whole log for every call.
func (s *Store) Header(id string) (Summary, error) {
	if id == "" {
		return Summary{}, fmt.Errorf("session id is required")
	}
	selected, err := s.lookupArtifact(id)
	if err != nil {
		return Summary{}, err
	}
	info, err := os.Lstat(selected.path)
	if err != nil {
		return Summary{}, err
	}
	if !info.Mode().IsRegular() {
		return Summary{}, fmt.Errorf("selected artifact is not a regular file")
	}
	headerCache.Lock()
	entry, ok := headerCache.byPath[selected.path]
	headerCache.Unlock()
	if ok && entry.summary.ID == id && info.Size() >= entry.size {
		// The header row cannot change while the log only grows, but the size
		// is a property of the file, not of the header.
		entry.summary.SizeBytes = info.Size()
		return entry.summary, nil
	}
	inspection, err := readArtifactHeader(selected)
	if err != nil {
		return Summary{}, fmt.Errorf("session %q (%s): %w", id, selected.path, err)
	}
	if inspection.ID != id {
		return Summary{}, fmt.Errorf("session header identity changed while reading %s", selected.path)
	}
	inspection.SizeBytes = info.Size()
	headerCache.Lock()
	headerCache.byPath[selected.path] = headerEntry{size: info.Size(), summary: inspection.Summary}
	headerCache.Unlock()
	return inspection.Summary, nil
}
