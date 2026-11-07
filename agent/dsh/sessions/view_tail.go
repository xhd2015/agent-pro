package sessions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type viewTailCursor struct {
	selected  artifact
	path      string
	size      int64
	offset    int64
	summary   Summary
	events    []Event
	view      *View
	truncated bool
	header    bool
	resyncSeq bool
	nextSeq   int64
}

type viewTailFlight struct {
	done chan struct{}
	view *View
	err  error
}

var viewTails = struct {
	mu       sync.Mutex
	cache    map[string]*viewTailCursor
	inflight map[string]*viewTailFlight
}{
	cache:    map[string]*viewTailCursor{},
	inflight: map[string]*viewTailFlight{},
}

// ViewTailReads counts payload decodes (cold suffix or append). Tests use it
// to prove an unchanged file is not re-read.
var ViewTailReads atomic.Int64

func resetViewTails() {
	viewTails.mu.Lock()
	viewTails.cache = map[string]*viewTailCursor{}
	viewTails.inflight = map[string]*viewTailFlight{}
	viewTails.mu.Unlock()
	ViewTailReads.Store(0)
}

// ViewTail returns a display projection of the newest keep messages without
// decoding the whole log. keep is a projected-row target (default 50).
func (s *Store) ViewTail(id string, keep int) (*View, error) {
	if keep <= 0 {
		keep = 50
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
	size := info.Size()
	key := selected.path
	viewTails.mu.Lock()
	cur := viewTails.cache[key]
	if cur != nil && cur.path == selected.path && cur.size == size && cur.view != nil {
		view := cur.view
		viewTails.mu.Unlock()
		return view, nil
	}
	if cur != nil && (cur.path != selected.path || size < cur.size) {
		cur = nil
	}
	if flight, ok := viewTails.inflight[key]; ok {
		viewTails.mu.Unlock()
		<-flight.done
		return flight.view, flight.err
	}
	flight := &viewTailFlight{done: make(chan struct{})}
	viewTails.inflight[key] = flight
	viewTails.mu.Unlock()

	var view *View
	var loadErr error
	if cur == nil {
		built, err := coldViewTail(selected, size, keep)
		loadErr = err
		if err == nil {
			viewTails.mu.Lock()
			viewTails.cache[key] = built
			viewTails.mu.Unlock()
			view = built.view
		}
	} else {
		loadErr = cur.appendTo(selected, size, keep)
		if loadErr == nil {
			viewTails.mu.Lock()
			viewTails.cache[key] = cur
			viewTails.mu.Unlock()
			view = cur.view
		}
	}
	viewTails.mu.Lock()
	flight.view = view
	flight.err = loadErr
	delete(viewTails.inflight, key)
	close(flight.done)
	viewTails.mu.Unlock()
	return view, loadErr
}

func coldViewTail(a artifact, size int64, keep int) (*viewTailCursor, error) {
	ViewTailReads.Add(1)
	cur := &viewTailCursor{
		selected: a,
		path:     a.path,
		summary:  Summary{Path: a.path, Encoding: a.encoding, FormatVersion: int(a.version), Outcome: "empty"},
	}
	if a.encoding == "zstd" && size > coldTailMin {
		if err := cur.coldZstd(a, size, keep); err != nil {
			return nil, err
		}
	} else {
		if err := cur.readRange(a, 0, size, false); err != nil {
			return nil, err
		}
		cur.truncated = false
	}
	cur.size = size
	cur.rebuild(keep)
	return cur, nil
}

func (c *viewTailCursor) coldZstd(a artifact, size int64, keep int) error {
	base, err := recentSummary(a)
	if err != nil {
		return err
	}
	c.summary = base
	c.header = true
	window := int64(coldTailWindow)
	for {
		if window > size {
			window = size
		}
		start, ok, err := suffixFrameStart(a.path, size, window)
		if err != nil {
			return err
		}
		if ok && start > 0 {
			probe := &viewTailCursor{
				selected:  a,
				path:      a.path,
				offset:    start,
				summary:   base,
				header:    true,
				resyncSeq: true,
				truncated: true,
			}
			if err := probe.readRange(a, start, size, true); err != nil {
				return err
			}
			if len(Project(probe.events)) >= keep || window == size {
				*c = *probe
				if window == size {
					c.truncated = false
				}
				return nil
			}
		}
		if window == size {
			break
		}
		window *= 2
	}
	c.offset = 0
	c.header = false
	c.resyncSeq = false
	c.truncated = false
	c.events = nil
	return c.readRange(a, 0, size, true)
}

func (c *viewTailCursor) appendTo(a artifact, size int64, keep int) error {
	if size <= c.size {
		c.size = size
		c.rebuild(keep)
		return nil
	}
	ViewTailReads.Add(1)
	if err := c.readRange(a, c.offset, size, a.encoding == "zstd"); err != nil {
		return err
	}
	c.size = size
	c.rebuild(keep)
	return nil
}

func (c *viewTailCursor) rebuild(keep int) {
	c.summary.SizeBytes = c.size
	view := ProjectInspection(&Inspection{Summary: c.summary, Events: c.events})
	view.Truncated = c.truncated
	c.view = view
	_ = keep
}

func (c *viewTailCursor) readRange(a artifact, from, size int64, zstdEnc bool) error {
	if size <= from {
		c.offset = size
		return nil
	}
	f, err := openArtifact(a.path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, size-from)
	n, err := f.ReadAt(buf, from)
	if err != nil && err != io.EOF {
		return err
	}
	buf = buf[:n]
	if zstdEnc {
		return c.consumeZstd(buf)
	}
	consumed, err := c.consumeJSONL(buf, false)
	c.offset += int64(consumed)
	return err
}

func (c *viewTailCursor) consumeZstd(suffix []byte) error {
	frames, tornStart, torn, err := scanZstdFrames(suffix)
	if err != nil {
		return err
	}
	for _, frame := range frames {
		plain, err := decodeZstdFrame(suffix[frame.start:frame.end])
		if err != nil {
			return err
		}
		if _, err := c.consumeJSONL(plain, true); err != nil {
			return err
		}
	}
	if torn {
		c.offset += int64(tornStart)
		return nil
	}
	c.offset += int64(len(suffix))
	return nil
}

func (c *viewTailCursor) consumeJSONL(data []byte, complete bool) (int, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	consumed := 0
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			if len(line) == 0 {
				return consumed, nil
			}
			if !complete {
				return consumed, nil
			}
			return consumed, fmt.Errorf("incomplete JSONL row")
		}
		if err != nil {
			return consumed, err
		}
		consumed += len(line)
		if err := c.consumeLine(bytes.TrimSpace(line)); err != nil {
			return consumed, err
		}
	}
}

func (c *viewTailCursor) consumeLine(line []byte) error {
	if len(line) == 0 {
		return nil
	}
	if !c.header {
		sum, err := parseSessionHeader(line, c.summary.FormatVersion, c.path, c.summary.Encoding)
		if err != nil {
			return err
		}
		sum.SizeBytes = c.summary.SizeBytes
		c.summary = sum
		c.header = true
		return nil
	}
	if c.resyncSeq {
		seq, err := lineSeq(line)
		if err != nil {
			return err
		}
		c.nextSeq = seq
		c.resyncSeq = false
	}
	event, err := parseEvent(line, c.nextSeq)
	if err != nil {
		return fmt.Errorf("event row %d: %w", c.nextSeq+1, err)
	}
	c.nextSeq++
	if err := project(&c.summary, event); err != nil {
		return err
	}
	if event.Time > c.summary.UpdatedAt {
		c.summary.UpdatedAt = event.Time
	}
	c.events = append(c.events, event)
	return nil
}
