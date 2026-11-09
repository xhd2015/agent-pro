package consumers

import (
	"sync"
	"testing"
	"time"
)

func TestPeekAnswersWithoutACache(t *testing.T) {
	f := newFixture(t)
	f.plain("session-peek", headerAt("session-peek", 4)+
		row("session/title", 1, `{"title":"Hello"}`))
	bus := f.bus(Options{})

	if snapshot, ok := bus.Peek("session-peek"); ok {
		t.Fatalf("peek before any sync returned %+v", snapshot)
	}
	if snapshot, ok := bus.Peek(""); ok {
		t.Fatalf("peek of an empty id returned %+v", snapshot)
	}
}

func TestPeekReturnsWhatSyncAlreadyFolded(t *testing.T) {
	f := newFixture(t)
	f.plain("session-peek", headerAt("session-peek", 4)+
		row("session/title", 1, `{"title":"Hello"}`)+
		row("user/message", 2, `{"content":[{"type":"text","text":"hi"}]}`))
	bus := f.bus(Options{})

	synced, err := bus.Sync("session-peek")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	peeked, ok := bus.Peek("session-peek")
	if !ok {
		t.Fatal("peek after sync missed the cached state")
	}
	if !peeked.CaughtUp || peeked.LastSeq != synced.LastSeq || peeked.Size != synced.Size {
		t.Fatalf("peeked = %+v, synced = %+v", peeked, synced)
	}
	if got := peeked.Title(); got != "Hello" {
		t.Fatalf("title = %q", got)
	}
	if got := peeked.Stats(); got.Turns != 1 {
		t.Fatalf("stats = %+v", got)
	}
	if peeked.ReadBytes != 0 || peeked.Folded != 0 {
		t.Fatalf("peek did work: %+v", peeked)
	}
}

// Peek never reads the log, so a session that grew since the last sync reports
// the old position with CaughtUp still true. Callers compare Size with the
// artifact to tell "folded to the end" from "folded to the end of what I read".
func TestPeekDoesNotReadTheLog(t *testing.T) {
	f := newFixture(t)
	path := f.plain("session-peek", headerAt("session-peek", 4)+
		row("session/title", 1, `{"title":"Hello"}`))
	bus := f.bus(Options{})
	if _, err := bus.Sync("session-peek"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	before := bus.Stats()

	appendBytes(t, path, []byte(row("session/title", 2, `{"title":"Changed"}`)))

	peeked, ok := bus.Peek("session-peek")
	if !ok {
		t.Fatal("peek missed the cached state")
	}
	if got := peeked.Title(); got != "Hello" {
		t.Fatalf("title = %q, want the folded one", got)
	}
	after := bus.Stats()
	if after.Events != before.Events || after.Folded != before.Folded || after.BytesRead != before.BytesRead {
		t.Fatalf("peek read the log: %+v -> %+v", before, after)
	}

	synced, err := bus.Sync("session-peek")
	if err != nil {
		t.Fatalf("resync: %v", err)
	}
	if got := synced.Title(); got != "Changed" {
		t.Fatalf("title after sync = %q", got)
	}
	if synced.LastSeq <= peeked.LastSeq {
		t.Fatalf("sync did not advance: %+v after %+v", synced, peeked)
	}
}

// A record written before a consumer existed cannot answer for it: its absent
// state would read as a real zero.
func TestPeekRejectsARecordMissingAConsumer(t *testing.T) {
	f := newFixture(t)
	f.plain("session-peek", headerAt("session-peek", 4)+
		row("session/title", 1, `{"title":"Hello"}`)+
		row("user/message", 2, `{"content":[{"type":"text","text":"hi"}]}`))

	narrow := f.bus(Options{Consumers: []Consumer{Title()}})
	if _, err := narrow.Sync("session-peek"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	wide := f.bus(Options{})
	if snapshot, ok := wide.Peek("session-peek"); ok {
		t.Fatalf("peek accepted a record without the stats fold: %+v", snapshot)
	}
	if _, err := wide.Sync("session-peek"); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if snapshot, ok := wide.Peek("session-peek"); !ok || snapshot.Stats().Turns != 1 {
		t.Fatalf("peek after the wider sync = %+v ok=%v", snapshot, ok)
	}
}

// A poll that lands while a cold fold runs must not queue behind it: the fold
// holds the session lock for as long as the log takes to read.
func TestPeekDoesNotWaitForARunningSync(t *testing.T) {
	f := newFixture(t)
	f.plain("session-peek", headerAt("session-peek", 4)+
		row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`))
	slow := &blockingConsumer{entered: make(chan struct{}), release: make(chan struct{})}
	bus := f.bus(Options{Consumers: []Consumer{slow}})

	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := bus.Sync("session-peek"); err != nil {
			t.Errorf("sync: %v", err)
		}
	}()
	<-slow.entered

	type answer struct {
		snapshot Snapshot
		ok       bool
	}
	answered := make(chan answer, 1)
	go func() {
		snapshot, ok := bus.Peek("session-peek")
		answered <- answer{snapshot, ok}
	}()
	select {
	case got := <-answered:
		if got.ok {
			t.Fatalf("peek during the first fold answered %+v", got.snapshot)
		}
	case <-time.After(2 * time.Second):
		slow.letGo()
		t.Fatal("peek blocked behind the running fold")
	}

	slow.letGo()
	<-done
	peeked, ok := bus.Peek("session-peek")
	if !ok || !peeked.CaughtUp || peeked.LastSeq != 1 {
		t.Fatalf("peek after the fold = %+v, %v", peeked, ok)
	}
	if peeked.ReadBytes != 0 || peeked.Folded != 0 {
		t.Fatalf("peek reported the fold's work: %+v", peeked)
	}
}

// blockingConsumer stalls one fold until the test releases it.
type blockingConsumer struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	onceRel sync.Once
}

func (c *blockingConsumer) letGo() { c.onceRel.Do(func() { close(c.release) }) }

func (c *blockingConsumer) ID() string      { return "blocking" }
func (c *blockingConsumer) Version() int    { return 1 }
func (c *blockingConsumer) Types() []string { return []string{"user/message"} }
func (c *blockingConsumer) New() any        { return new(int64) }

func (c *blockingConsumer) Fold(state any, ev Event, at Point) error {
	c.once.Do(func() { close(c.entered) })
	<-c.release
	return nil
}

func TestPeekMissesAfterForget(t *testing.T) {
	f := newFixture(t)
	f.plain("session-peek", headerAt("session-peek", 4)+
		row("session/title", 1, `{"title":"Hello"}`))
	bus := f.bus(Options{})
	mustSync(t, bus, "session-peek")
	if _, ok := bus.Peek("session-peek"); !ok {
		t.Fatal("peek after sync missed")
	}
	if err := bus.Forget("session-peek"); err != nil {
		t.Fatalf("forget: %v", err)
	}
	if snapshot, ok := bus.Peek("session-peek"); ok {
		t.Fatalf("peek after forget answered %+v", snapshot)
	}
}
