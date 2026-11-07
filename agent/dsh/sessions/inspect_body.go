package sessions

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// inspectBodyCap bounds decoded snapshots by count. Event payloads are also
// bounded by inspectBodyEventBytesCap (artifact bytes whose Events are retained).
const inspectBodyCap = 16

// inspectBodyEventBytesCap is compressed artifact bytes kept with Events.
// View drops Events after projecting; a 32MiB budget is a few large logs, not 16.
var inspectBodyEventBytesCap int64 = 32 << 20

type inspectBody struct {
	selected   artifact
	path       string
	size       int64
	modNano    int64
	inspection *Inspection
	view       *View
	used       time.Time
}

type inspectBodyFlight struct {
	done chan struct{}
	body *inspectBody
	err  error
}

var inspectBodies = struct {
	mu       sync.Mutex
	cache    map[string]*inspectBody
	inflight map[string]*inspectBodyFlight
}{
	cache:    map[string]*inspectBody{},
	inflight: map[string]*inspectBodyFlight{},
}

// InspectReads counts actual readArtifact calls (cache misses). Tests use it
// to prove an unchanged file is not re-decoded.
var InspectReads atomic.Int64

func resetInspectBodies() {
	inspectBodies.mu.Lock()
	inspectBodies.cache = map[string]*inspectBody{}
	inspectBodies.inflight = map[string]*inspectBodyFlight{}
	inspectBodies.mu.Unlock()
	InspectReads.Store(0)
}

// ResetInspectCache drops decoded Inspect/View snapshots. Tests use it to
// isolate load counts; production callers should not need it.
func ResetInspectCache() {
	resetInspectBodies()
	resetArtifactIDs()
	inspectMemo.mu.Lock()
	inspectMemo.index = nil
	inspectMemo.mu.Unlock()
}

func (s *Store) loadInspectBody(id string) (*inspectBody, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is required")
	}
	selected, err := s.lookupArtifact(id)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(selected.path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("selected artifact is not a regular file")
	}

	key := selected.path
	inspectBodies.mu.Lock()
	if cached, ok := inspectBodies.cache[key]; ok && inspectBodyFresh(cached, info) {
		cached.used = time.Now()
		inspectBodies.mu.Unlock()
		if cached.inspection.ID != id {
			return nil, fmt.Errorf("session header identity changed while reading %s", selected.path)
		}
		return cached, nil
	}
	if flight, ok := inspectBodies.inflight[key]; ok {
		inspectBodies.mu.Unlock()
		<-flight.done
		if flight.err != nil {
			return nil, flight.err
		}
		if flight.body.inspection.ID != id {
			return nil, fmt.Errorf("session header identity changed while reading %s", selected.path)
		}
		return flight.body, nil
	}
	flight := &inspectBodyFlight{done: make(chan struct{})}
	inspectBodies.inflight[key] = flight
	inspectBodies.mu.Unlock()

	body, err := readInspectBody(selected, id)

	inspectBodies.mu.Lock()
	if err == nil {
		body.used = time.Now()
		inspectBodies.cache[key] = body
		evictInspectBodiesLocked()
	} else {
		delete(inspectBodies.cache, key)
	}
	flight.body = body
	flight.err = err
	delete(inspectBodies.inflight, key)
	close(flight.done)
	inspectBodies.mu.Unlock()
	return body, err
}

func inspectBodyFresh(body *inspectBody, info os.FileInfo) bool {
	return body != nil && info.Size() == body.size && info.ModTime().UnixNano() == body.modNano
}

func readInspectBody(selected artifact, id string) (*inspectBody, error) {
	InspectReads.Add(1)
	result, err := readArtifact(selected)
	if err != nil {
		return nil, fmt.Errorf("session %q (%s): %w", id, selected.path, err)
	}
	if result.ID != id {
		return nil, fmt.Errorf("session header identity changed while reading %s", selected.path)
	}
	info, err := os.Lstat(selected.path)
	if err != nil {
		return nil, err
	}
	return &inspectBody{
		selected:   selected,
		path:       selected.path,
		size:       info.Size(),
		modNano:    info.ModTime().UnixNano(),
		inspection: result,
	}, nil
}

func (body *inspectBody) holdsEvents() bool {
	return body != nil && body.inspection != nil && body.inspection.Events != nil
}

func inspectEventBytesLocked() int64 {
	var n int64
	for _, body := range inspectBodies.cache {
		if body.holdsEvents() {
			n += body.size
		}
	}
	return n
}

func dropInspectEventsLocked(body *inspectBody) {
	if body != nil && body.inspection != nil {
		body.inspection.Events = nil
	}
}

func (s *Store) ensureInspectEvents(body *inspectBody) error {
	if body == nil {
		return fmt.Errorf("session inspect body is missing")
	}
	inspectBodies.mu.Lock()
	if body.holdsEvents() {
		inspectBodies.mu.Unlock()
		return nil
	}
	selected := body.selected
	id := ""
	if body.inspection != nil {
		id = body.inspection.ID
	}
	inspectBodies.mu.Unlock()
	fresh, err := readInspectBody(selected, id)
	if err != nil {
		return err
	}
	inspectBodies.mu.Lock()
	body.inspection = fresh.inspection
	body.size = fresh.size
	body.modNano = fresh.modNano
	body.selected = fresh.selected
	evictInspectBodiesLocked()
	inspectBodies.mu.Unlock()
	return nil
}

func evictInspectBodiesLocked() {
	for len(inspectBodies.cache) > inspectBodyCap || inspectEventBytesLocked() > inspectBodyEventBytesCap {
		var oldestKey string
		var oldest time.Time
		first := true
		for key, body := range inspectBodies.cache {
			if first || body.used.Before(oldest) {
				oldestKey = key
				oldest = body.used
				first = false
			}
		}
		if oldestKey == "" {
			return
		}
		delete(inspectBodies.cache, oldestKey)
	}
}
