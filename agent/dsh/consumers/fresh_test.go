package consumers

import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// Fresh is the difference between "the fold reached the end of what I read" and
// "what I read is still the log": an append leaves CaughtUp true and Fresh false.
func TestPeekFreshOnlyWhileTheLogIsUnchanged(t *testing.T) {
	f := newFixture(t)
	path := f.plain("session-fresh", headerAt("session-fresh", 4)+
		row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`))
	bus := f.bus(Options{})

	synced, err := bus.Sync("session-fresh")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !synced.Fresh {
		t.Fatalf("a finished sync of an unchanged log is not fresh: %+v", synced)
	}
	if synced.Modified.IsZero() {
		t.Fatal("a fresh snapshot must report the artifact's last write")
	}

	peeked, ok := bus.Peek("session-fresh")
	if !ok || !peeked.Fresh {
		t.Fatalf("peek after sync = %+v ok=%v, want fresh", peeked, ok)
	}

	appendBytes(t, path, []byte(row("assistant/message", 2,
		`{"message":{"content":[{"type":"text","text":"yo"}]}}`)))

	stale, ok := bus.Peek("session-fresh")
	if !ok {
		t.Fatal("peek missed the cached state")
	}
	if stale.Fresh {
		t.Fatalf("peek called a snapshot fresh after the log grew: %+v", stale)
	}
	if !stale.CaughtUp {
		t.Fatalf("CaughtUp describes what was read, not the current file: %+v", stale)
	}
	if !stale.Modified.After(synced.Modified) {
		t.Fatalf("modified did not advance: %v -> %v", synced.Modified, stale.Modified)
	}

	again, err := bus.Sync("session-fresh")
	if err != nil {
		t.Fatalf("resync: %v", err)
	}
	if !again.Fresh {
		t.Fatalf("resync did not make the snapshot fresh: %+v", again)
	}
}

// A record is keyed by session id, and two homes can hold the same id. Peek must
// refuse a record whose artifact this store does not resolve, or a pinned backend
// would paint the other home's fold.
func TestPeekRejectsARecordFromAnotherRoot(t *testing.T) {
	f := newFixture(t)
	body := headerAt("session-foreign", 4) +
		row("user/message", 1, `{"content":[{"type":"text","text":"hi"}]}`)
	f.plain("session-foreign", body)
	if _, err := f.bus(Options{}).Sync("session-foreign"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	otherRoot := t.TempDir()
	writeArtifact(t, otherRoot, "session-foreign", "session.v3.jsonl", body)
	other := &sessions.Store{Root: otherRoot}
	bus, err := New(other, Options{Dir: f.cache, Consumers: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot, ok := bus.Peek("session-foreign"); ok {
		t.Fatalf("peek accepted a record for another home's artifact: %+v", snapshot)
	}
	snapshot, err := bus.Sync("session-foreign")
	if err != nil {
		t.Fatalf("sync under the other root: %v", err)
	}
	if snapshot.ArtifactPath == "" || !snapshot.Fresh {
		t.Fatalf("the new home's own read = %+v", snapshot)
	}
}

// Home keeps the records beside the home they describe, so two backends never
// share one.
func TestOptionsHomeKeepsRecordsBesideTheHome(t *testing.T) {
	home := t.TempDir()
	bus, err := New(&sessions.Store{Root: t.TempDir()}, Options{Home: home, Consumers: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "consumers"); bus.CacheDir() != want {
		t.Fatalf("cache dir = %q, want %q", bus.CacheDir(), want)
	}
}

// The retry decision needs the failure's own time: a time stamped by the reader
// would make every poll look like a new failure.
func TestOutcomeFailureCarriesTheRowTime(t *testing.T) {
	f := newFixture(t)
	f.plain("session-failure", headerAt("session-failure", 4)+
		row("turn/end", 1, `{"turn":0,"reason":{"kind":"error","error":{"message":"boom","code":"SERVER"}}}`)+
		row("llm/retry", 2, `{"failure":{"message":"again","code":"TIMEOUT"}}`))
	snapshot, err := f.bus(Options{}).Sync("session-failure")
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	failure := snapshot.Failure()
	if failure == nil {
		t.Fatal("no failure folded")
	}
	if failure.Message != "again" || failure.Code != "TIMEOUT" {
		t.Fatalf("failure = %+v, want the last one", failure)
	}
	// row() stamps time as 20+seq.
	if failure.At != 22 || failure.Seq != 2 {
		t.Fatalf("failure identity = seq %d at %d, want 2/22", failure.Seq, failure.At)
	}
}

// Dropping one consumer's cached state is how a fold whose rules changed gets
// replayed: the record no longer answers, and the next Sync rebuilds it from the
// stream origin while the surviving consumers skip what they already folded.
func TestDropConsumerForcesAReplay(t *testing.T) {
	f := newFixture(t)
	f.plain("session-drop", headerAt("session-drop", 4)+
		row("session/title", 1, `{"title":"Hello"}`)+
		row("user/message", 2, `{"content":[{"type":"text","text":"hi"}]}`))
	bus := f.bus(Options{})
	if _, err := bus.Sync("session-drop"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	scanned, changed, err := DropConsumer(f.cache, TitleID, true)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if scanned != 1 || changed != 1 {
		t.Fatalf("dry run scanned=%d changed=%d", scanned, changed)
	}

	scanned, changed, err = DropConsumer(f.cache, TitleID, false)
	if err != nil {
		t.Fatalf("drop: %v", err)
	}
	if scanned != 1 || changed != 1 {
		t.Fatalf("scanned=%d changed=%d", scanned, changed)
	}

	// A new process over the same cache sees an uncovered record: no answer.
	fresh, err := New(f.store, Options{Dir: f.cache, Consumers: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot, ok := fresh.Peek("session-drop"); ok {
		t.Fatalf("a record missing a consumer answered: %+v", snapshot)
	}
	replayed, err := fresh.Sync("session-drop")
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if got := replayed.Title(); got != "Hello" {
		t.Fatalf("replayed title = %q", got)
	}
	if !replayed.Fresh {
		t.Fatalf("replay is not fresh: %+v", replayed)
	}
}
