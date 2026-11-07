package sessions

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// viewTailsMaxEntries bounds retained cursors. A cursor keeps the projected
// rows of its suffix window, so this bounds the memory a burst of session
// previews can pin (the map used to grow without limit).
const viewTailsMaxEntries = 8

// beforeSeqGrowSteps bounds how many window doublings one earlier-page request
// may decode, so a click costs a bounded amount of work.
const beforeSeqGrowSteps = 4

type viewTailCursor struct {
	selected  artifact
	path      string
	size      int64
	offset    int64
	start     int64 // file offset the retained rows begin at; 0 = whole log
	summary   Summary
	view      *View
	truncated bool
	header    bool
	resyncSeq bool
	nextSeq   int64
	lastUsed  int64
	// rows holds the projection folded in event by event; rebuild only
	// assembles a View from it, so an append never re-projects the window.
	rows       []Message
	byCall     map[string]*Ask
	byCallAt   map[string]int
	answers    map[string][]AskAnswer
	todos      []Todo
	effort     string
	sawOutcome bool // this window holds a turn event, so its Outcome is final
	pinned     bool // rows backing array is referenced by a published View
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

// viewTailClock orders cursor use for the cache's least-recently-used eviction.
var viewTailClock atomic.Int64

// ViewTailReads counts payload decodes (cold suffix, append, window growth).
// Tests use it to prove an unchanged file is not re-read.
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
	want := func(c *viewTailCursor) bool { return len(c.rows) >= keep }
	return s.viewTail(id, keep, want, false)
}

// ViewTailBefore returns the projection that holds at least keep rows older
// than beforeSeq, widening the decoded suffix backwards when the cached window
// is too narrow. It replaces the full-log re-decode that every earlier page
// used to pay on logs past the inspection cache cap.
func (s *Store) ViewTailBefore(id string, beforeSeq int64, keep int, eligible func(Message) bool) (*View, error) {
	if beforeSeq <= 0 {
		return s.ViewTail(id, keep)
	}
	if keep <= 0 {
		keep = 50
	}
	if eligible == nil {
		eligible = func(Message) bool { return true }
	}
	want := func(c *viewTailCursor) bool { return c.olderRows(beforeSeq, eligible) >= keep }
	return s.viewTail(id, keep, want, true)
}

func (s *Store) viewTail(id string, keep int, want func(*viewTailCursor) bool, windowed bool) (*View, error) {
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
	for attempt := 0; ; attempt++ {
		viewTails.mu.Lock()
		cur := viewTails.cache[key]
		if cur != nil && (cur.path != selected.path || size < cur.size) {
			cur = nil
		}
		// A loader outside the lock mutates the cached cursor, so only a key
		// with no in-flight load may be inspected or served here.
		if flight, busy := viewTails.inflight[key]; busy {
			viewTails.mu.Unlock()
			<-flight.done
			// A concurrent load may have kept a narrower window than this
			// caller needs; retry so the request still gets its older rows.
			if attempt >= 1 {
				return flight.view, flight.err
			}
			continue
		}
		if cur != nil && cur.size == size && cur.view != nil && (!windowed || want(cur)) {
			cur.lastUsed = viewTailClock.Add(1)
			view := cur.view
			viewTails.mu.Unlock()
			return view, nil
		}
		flight := &viewTailFlight{done: make(chan struct{})}
		viewTails.inflight[key] = flight
		viewTails.mu.Unlock()

		var built *viewTailCursor
		var loadErr error
		if cur == nil {
			built, loadErr = coldViewTail(selected, size, keep, want)
		} else {
			built = cur
			if built.size != size {
				loadErr = built.appendTo(selected, size, keep)
			}
			if loadErr == nil && windowed && !want(built) {
				_, loadErr = built.growBackward(selected, size, want)
				if loadErr == nil {
					built.size = size
					built.rebuild(keep)
				}
			}
		}
		var view *View
		if loadErr == nil {
			viewTails.mu.Lock()
			storeViewTailLocked(key, built)
			viewTails.mu.Unlock()
			view = built.view
		} else {
			// A half-applied cursor must not be reused: drop it so the next
			// request rebuilds from the file instead of double-appending.
			viewTails.mu.Lock()
			if viewTails.cache[key] == built {
				delete(viewTails.cache, key)
			}
			viewTails.mu.Unlock()
		}
		viewTails.mu.Lock()
		flight.view = view
		flight.err = loadErr
		delete(viewTails.inflight, key)
		close(flight.done)
		viewTails.mu.Unlock()
		return view, loadErr
	}
}

func storeViewTailLocked(key string, cur *viewTailCursor) {
	cur.lastUsed = viewTailClock.Add(1)
	viewTails.cache[key] = cur
	for len(viewTails.cache) > viewTailsMaxEntries {
		victim := ""
		oldest := int64(0)
		for other, cached := range viewTails.cache {
			if other == key {
				continue
			}
			if victim == "" || cached.lastUsed < oldest {
				victim, oldest = other, cached.lastUsed
			}
		}
		if victim == "" {
			return
		}
		delete(viewTails.cache, victim)
	}
}

func coldViewTail(a artifact, size int64, keep int, want func(*viewTailCursor) bool) (*viewTailCursor, error) {
	ViewTailReads.Add(1)
	cur := &viewTailCursor{
		selected: a,
		path:     a.path,
		summary:  Summary{Path: a.path, Encoding: a.encoding, FormatVersion: int(a.version), Outcome: "empty"},
	}
	if a.encoding == "zstd" && size > coldTailMin {
		if err := cur.coldZstd(a, size, want); err != nil {
			return nil, err
		}
	} else {
		// Small zstd logs are decoded whole; the encoding still decides whether
		// the bytes are framed or plain JSONL.
		if err := cur.readRange(a, 0, size, a.encoding == "zstd"); err != nil {
			return nil, err
		}
		cur.truncated = false
	}
	cur.size = size
	cur.rebuild(keep)
	return cur, nil
}

func (c *viewTailCursor) coldZstd(a artifact, size int64, want func(*viewTailCursor) bool) error {
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
				start:     start,
				summary:   base,
				header:    true,
				resyncSeq: true,
				truncated: true,
			}
			if err := probe.readRange(a, start, size, true); err != nil {
				return err
			}
			if want(probe) || window == size {
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
		if next := window * 2; next > coldTailMaxWindow && size > next {
			break
		}
		window *= 2
	}
	c.offset = 0
	c.start = 0
	c.header = false
	c.resyncSeq = false
	c.truncated = false
	c.rows = nil
	c.byCall = nil
	c.byCallAt = nil
	c.answers = nil
	c.todos = nil
	c.effort = ""
	c.sawOutcome = false
	c.nextSeq = 0
	return c.readRange(a, 0, size, true)
}

// growBackward widens the decoded suffix to an earlier frame boundary and
// prepends only the newly exposed rows, so paging to earlier messages costs the
// bytes it adds instead of a whole-log decode.
func (c *viewTailCursor) growBackward(a artifact, size int64, want func(*viewTailCursor) bool) (bool, error) {
	window := size - c.start
	if window <= 0 {
		window = int64(coldTailWindow)
	}
	for step := 0; step < beforeSeqGrowSteps; step++ {
		if c.start <= 0 {
			break
		}
		if want(c) {
			return true, nil
		}
		if a.encoding != "zstd" {
			break
		}
		window *= 2
		start, ok, err := suffixFrameStart(a.path, size, window)
		if err != nil {
			return false, err
		}
		if !ok || start >= c.start {
			break
		}
		base, err := recentSummary(a)
		if err != nil {
			return false, err
		}
		probe := &viewTailCursor{
			selected:  a,
			path:      a.path,
			offset:    start,
			start:     start,
			summary:   base,
			header:    true,
			resyncSeq: true,
			truncated: true,
		}
		atHead := start == 0
		if atHead {
			// The window now reaches the log head, so this segment starts
			// with the session header line.
			probe.header = false
			probe.resyncSeq = false
			probe.truncated = false
		}
		ViewTailReads.Add(1)
		if err := probe.readRange(a, start, c.start, true); err != nil {
			return false, err
		}
		c.prepend(probe)
		if atHead {
			c.truncated = false
			break
		}
	}
	return want(c), nil
}

// prepend merges an earlier segment into c. The result rows are a fresh slice:
// previously published Views keep their own window untouched.
func (c *viewTailCursor) prepend(earlier *viewTailCursor) {
	offset := len(earlier.rows)
	rows := make([]Message, 0, offset+len(c.rows))
	rows = append(rows, earlier.rows...)
	rows = append(rows, c.rows...)
	if earlier.byCall != nil || c.byCall != nil {
		byCall := make(map[string]*Ask, len(earlier.byCall)+len(c.byCall))
		byCallAt := make(map[string]int, len(earlier.byCall)+len(c.byCall))
		for callID, ask := range earlier.byCall {
			byCall[callID] = ask
			byCallAt[callID] = earlier.byCallAt[callID]
		}
		for callID, ask := range c.byCall {
			byCall[callID] = ask
			byCallAt[callID] = offset + c.byCallAt[callID]
		}
		c.byCall = byCall
		c.byCallAt = byCallAt
	}
	answers := c.answers
	if earlier.answers != nil {
		merged := make(map[string][]AskAnswer, len(earlier.answers)+len(c.answers))
		for callID, list := range earlier.answers {
			merged[callID] = list
		}
		for callID, list := range c.answers {
			merged[callID] = list
		}
		answers = merged
	}
	c.answers = answers
	c.rows = rows
	c.pinned = false
	// The answer row sits in the newer segment while its question is now in the
	// earlier one, so the prepended ask rows need the known answers reapplied.
	for callID, list := range answers {
		at, ok := c.byCallAt[callID]
		if !ok || at >= len(rows) || rows[at].Ask == nil {
			continue
		}
		clone := *rows[at].Ask
		clone.Answers = list
		rows[at].Ask = &clone
		c.byCall[callID] = &clone
	}
	c.start = earlier.start
	if c.effort == "" {
		c.effort = earlier.effort
	}
	if c.todos == nil {
		c.todos = earlier.todos
	}
	if !c.sawOutcome {
		c.summary.Outcome = earlier.summary.Outcome
		c.sawOutcome = earlier.sawOutcome
	}
	if c.summary.Title == "" {
		c.summary.Title = earlier.summary.Title
	}
	if earlier.summary.UpdatedAt > c.summary.UpdatedAt {
		c.summary.UpdatedAt = earlier.summary.UpdatedAt
	}
	if earlier.summary.FormatVersion != 0 && c.summary.FormatVersion == 0 {
		c.summary.FormatVersion = earlier.summary.FormatVersion
	}
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

// rebuild only assembles the cached projection; the per-event folding happened
// in applyEvent, so a poll over a growing log stays O(1).
func (c *viewTailCursor) rebuild(keep int) {
	c.summary.SizeBytes = c.size
	c.pinned = true
	c.view = &View{
		Summary:         c.summary,
		ReasoningEffort: c.effort,
		Todos:           c.todos,
		Rows:            c.rows,
		Truncated:       c.truncated,
	}
	_ = keep
}

// applyEvent folds one decoded event into the cursor's display state.
func (c *viewTailCursor) applyEvent(event Event) {
	if effort := eventReasoningEffort(event); effort != "" {
		c.effort = effort
	}
	if todos, ok := eventAgentTodos(event); ok {
		c.todos = todos
	}
	switch event.Type {
	case "turn/start", "turn/end":
		c.sawOutcome = true
	}
	if callID, answers := askResult(event); callID != "" && len(answers) > 0 {
		c.applyAnswer(callID, answers)
	}
	for _, row := range projectEvent(event) {
		if row.Ask != nil && row.Ask.CallID != "" {
			if answers, ok := c.answers[row.Ask.CallID]; ok {
				clone := *row.Ask
				clone.Answers = answers
				row.Ask = &clone
			}
			if c.byCall == nil {
				c.byCall = map[string]*Ask{}
				c.byCallAt = map[string]int{}
			}
			c.byCall[row.Ask.CallID] = row.Ask
			c.byCallAt[row.Ask.CallID] = len(c.rows)
		}
		c.rows = append(c.rows, row)
	}
}

// applyAnswer fills the ask row that owns callID. Published Views share the row
// array, so the write happens on a private copy and on a cloned Ask.
func (c *viewTailCursor) applyAnswer(callID string, answers []AskAnswer) {
	if c.answers == nil {
		c.answers = map[string][]AskAnswer{}
	}
	c.answers[callID] = answers
	at, ok := c.byCallAt[callID]
	if !ok || at >= len(c.rows) {
		return
	}
	ask := c.rows[at].Ask
	if ask == nil {
		return
	}
	clone := *ask
	clone.Answers = answers
	c.copyRowsBeforeWrite()
	c.rows[at].Ask = &clone
	if c.byCall != nil {
		c.byCall[callID] = &clone
	}
}

func (c *viewTailCursor) copyRowsBeforeWrite() {
	if !c.pinned {
		return
	}
	rows := make([]Message, len(c.rows))
	copy(rows, c.rows)
	c.rows = rows
	c.pinned = false
}

func (c *viewTailCursor) olderRows(beforeSeq int64, eligible func(Message) bool) int {
	count := 0
	for _, row := range c.rows {
		if row.Seq < beforeSeq && eligible(row) {
			count++
		}
	}
	return count
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

// consumeJSONL scans lines in place. data is a transient decode buffer, so
// lines are handed out as subslices instead of copies; every value that
// outlives this call is cloned by parseEvent.
func (c *viewTailCursor) consumeJSONL(data []byte, complete bool) (int, error) {
	consumed := 0
	for consumed < len(data) {
		rest := data[consumed:]
		idx := bytes.IndexByte(rest, '\n')
		if idx < 0 {
			if !complete {
				return consumed, nil
			}
			return consumed, fmt.Errorf("incomplete JSONL row")
		}
		consumed += idx + 1
		if err := c.consumeLine(bytes.TrimSpace(rest[:idx])); err != nil {
			return consumed, err
		}
	}
	return consumed, nil
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
	c.applyEvent(event)
	return nil
}
