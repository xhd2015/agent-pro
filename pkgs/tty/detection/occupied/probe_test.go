package occupied

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestProbe_exactlyOneSpaceIsOccupied(t *testing.T) {
	var mu sync.Mutex
	text := "draft"
	io := IO{
		Snapshot: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			return text, nil
		},
		Inject: func(s string) error {
			mu.Lock()
			defer mu.Unlock()
			switch s {
			case " ":
				text = text + " "
			case "\x7f":
				if len(text) > 0 && text[len(text)-1] == ' ' {
					text = text[:len(text)-1]
				}
			}
			return nil
		},
		Settle: time.Millisecond,
	}
	if got := Probe(io); got != Occupied {
		t.Fatalf("Probe=%q want occupied", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if text != "draft" {
		t.Fatalf("after DEL text=%q want draft", text)
	}
}

func TestProbe_placeholderCollapseIsEmpty(t *testing.T) {
	var mu sync.Mutex
	text := "Ask Codex to do anything"
	io := IO{
		Snapshot: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			return text, nil
		},
		Inject: func(s string) error {
			mu.Lock()
			defer mu.Unlock()
			switch s {
			case " ":
				text = "" // placeholder collapses
			case "\x7f":
				text = "Ask Codex to do anything"
			}
			return nil
		},
		Settle: time.Millisecond,
	}
	if got := Probe(io); got != Empty {
		t.Fatalf("Probe=%q want empty", got)
	}
}

func TestProbe_injectFailIsEmpty(t *testing.T) {
	io := IO{
		Snapshot: func() (string, error) { return "x", nil },
		Inject:   func(string) error { return errors.New("boom") },
	}
	if got := Probe(io); got != Empty {
		t.Fatalf("Probe=%q want empty (inject fail cannot prove occupied)", got)
	}
}

func TestProbe_beforeSkipsFirstSnapshot(t *testing.T) {
	snaps := 0
	io := IO{
		Before: "draft",
		Snapshot: func() (string, error) {
			snaps++
			return "draft ", nil
		},
		Inject: func(string) error { return nil },
		Settle: time.Millisecond,
	}
	if got := Probe(io); got != Occupied {
		t.Fatalf("Probe=%q want occupied", got)
	}
	if snaps != 1 {
		t.Fatalf("Snapshot calls=%d want 1 (after only)", snaps)
	}
}

func TestProbe_noChangeIsEmpty(t *testing.T) {
	io := IO{
		Snapshot: func() (string, error) { return "same", nil },
		Inject:   func(string) error { return nil },
		Settle:   time.Millisecond,
	}
	if got := Probe(io); got != Empty {
		t.Fatalf("Probe=%q want empty", got)
	}
}
