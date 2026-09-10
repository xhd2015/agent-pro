package idle

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSnapMetaStable(t *testing.T) {
	h1, n1 := SnapMeta("abc")
	h2, n2 := SnapMeta("abc")
	if h1 != h2 || n1 != 3 || n2 != 3 {
		t.Fatalf("hash/len unstable: %s/%d vs %s/%d", h1, n1, h2, n2)
	}
	h3, _ := SnapMeta("abd")
	if h3 == h1 {
		t.Fatal("different snap must different hash")
	}
}

func TestSnapTailTruncatesRunes(t *testing.T) {
	in := strings.Repeat("x", 50) + "TAIL"
	got := SnapTail(in, 4)
	if got != "TAIL" {
		t.Fatalf("SnapTail=%q want TAIL", got)
	}
}

func TestJSONL_appendsEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions", "sess-1", "idle.jsonl")
	j := &JSONL{Path: path}
	j.Log(Event{Event: EventArmed, SessionID: "sess-1", Timeout: "10m0s", TS: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	j.Log(Event{Event: EventTick, SessionID: "sess-1", Hits: 1, TS: time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC)})

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var events []Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("line %q: %v", sc.Text(), err)
		}
		events = append(events, e)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("lines=%d want 2", len(events))
	}
	if events[0].Event != EventArmed || events[1].Event != EventTick {
		t.Fatalf("events=%v", events)
	}
	if events[0].SessionID != "sess-1" || events[1].Hits != 1 {
		t.Fatalf("payload: %+v %+v", events[0], events[1])
	}
}
