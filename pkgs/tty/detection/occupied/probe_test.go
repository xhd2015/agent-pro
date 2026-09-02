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
	delDone := false
	io := IO{
		Before: "draft",
		Snapshot: func() (string, error) {
			snaps++
			// After DEL, return baseline so stabilize exits (Inject is a no-op).
			if delDone {
				return "draft", nil
			}
			return "draft ", nil
		},
		Inject: func(s string) error {
			if s == "\x7f" {
				delDone = true
			}
			return nil
		},
		Settle: time.Millisecond,
	}
	if got := Probe(io); got != Occupied {
		t.Fatalf("Probe=%q want occupied", got)
	}
	// One classify Snapshot after space; stabilize may Snapshot again.
	if snaps < 1 {
		t.Fatalf("Snapshot calls=%d want >=1 (classify after)", snaps)
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

// Live Codex crime scene (idle condition-1): space probe shortens the composer
// line and the visible snapshot scrolls — DEL restores "Ask Codex…" but does
// not bring back the scrolled-off head. Probe owns recovery: after it returns,
// Snapshot must equal the pre-probe resting text (keystroke undo is not enough).
func TestProbe_restoresRestingSnapshotAfterScrollSideEffect(t *testing.T) {
	const (
		prefix   = "i (alice)\n  | Team Role...(omitted)\n  message_id=u"
		body     = "7oirumAclAwbxQB-RCdFwv1vXNAhNTIDSuo6u8zN\n  [user@example.com]\n"
		composer = "› Ask Codex to do anything\n  gpt-5.6-luna max · ~/ws\n"
		footer   = ""
	)
	resting := prefix + body + composer
	// After space: head scrolled off (mid message_id) + placeholder collapsed.
	afterSpace := body + "›  \n  gpt-5.6-luna max · ~/ws\n"
	// After DEL: composer restored, viewport head still scrolled (incomplete undo).
	afterDEL := body + composer

	var mu sync.Mutex
	text := resting
	postDelSnaps := 0
	io := IO{
		Before: resting,
		Snapshot: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			// After DEL, first poll still scrolled; later polls fully restore
			// (Probe stabilize-to-before waits for this).
			if postDelSnaps > 0 {
				postDelSnaps++
				if postDelSnaps >= 2 {
					text = resting
				}
			}
			return text, nil
		},
		Inject: func(s string) error {
			mu.Lock()
			defer mu.Unlock()
			switch s {
			case " ":
				text = afterSpace
				postDelSnaps = 0
			case "\x7f":
				text = afterDEL
				postDelSnaps = 1
			}
			return nil
		},
		Settle: time.Millisecond,
	}
	if got := Probe(io); got != Empty {
		t.Fatalf("Probe=%q want empty (placeholder collapse is not occupy)", got)
	}
	mu.Lock()
	got := text
	mu.Unlock()
	if got != resting {
		t.Fatalf("Probe must restore pre-probe resting snapshot; got len=%d want len=%d\n got head=%q\nwant head=%q",
			len(got), len(resting), clipProbe(got, 60), clipProbe(resting, 60))
	}
}

func clipProbe(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
