package idle

import (
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
)

func TestWatchdog_threeEmptyStableSoftExit(t *testing.T) {
	var soft, shut int
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout:       10 * time.Second,
		Grace:         5 * time.Second,
		Now:           func() time.Time { return now },
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft++ },
		Shutdown:      func() { shut++ },
	})
	w.Tick() // t0 hit1
	now = t0.Add(5 * time.Second)
	w.Tick() // hit2
	now = t0.Add(10 * time.Second)
	w.Tick() // hit3 + Timeout elapsed → SoftExit
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1", soft)
	}
	now = t0.Add(15 * time.Second)
	w.Tick()
	if shut != 1 {
		t.Fatalf("Shutdown=%d want 1", shut)
	}
}

func TestWatchdog_changedResets(t *testing.T) {
	var soft int
	snap := "a"
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return snap, nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft++ },
	})
	w.Tick()
	w.Tick()
	snap = "b"
	w.Tick()
	if w.IdleHits() != 0 || soft != 0 {
		t.Fatalf("hits=%d soft=%d want 0,0 after change", w.IdleHits(), soft)
	}
}

func TestWatchdog_occupiedResets(t *testing.T) {
	var soft int
	status := occupied.Empty
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return status },
		SoftExit:      func() { soft++ },
	})
	w.Tick()
	w.Tick()
	status = occupied.Occupied
	w.Tick()
	if w.IdleHits() != 0 || soft != 0 {
		t.Fatalf("hits=%d soft=%d want 0,0 when occupied", w.IdleHits(), soft)
	}
}

func TestWatchdog_unknownAfterReadyCountsIdle(t *testing.T) {
	// After Ready, Unknown must not block exit (treat as empty).
	var soft int
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout:       10 * time.Millisecond,
		Now:           func() time.Time { return now },
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Unknown },
		SoftExit:      func() { soft++ },
	})
	w.Tick()
	if w.IdleHits() != 1 {
		t.Fatalf("hits=%d want 1 when Unknown after Ready", w.IdleHits())
	}
	now = t0.Add(5 * time.Millisecond)
	w.Tick()
	now = t0.Add(15 * time.Millisecond)
	w.Tick()
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1", soft)
	}
}

func TestWatchdog_queueHolds(t *testing.T) {
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		QueueLen:      func() int { return 1 },
	})
	w.Tick()
	if w.IdleHits() != 0 {
		t.Fatalf("hits=%d want 0 when queue non-zero", w.IdleHits())
	}
}

func TestWatchdog_finalOccupyCheckHoldsSoftExit(t *testing.T) {
	var soft int
	n := 0
	statuses := []occupied.Status{occupied.Empty, occupied.Empty, occupied.Empty, occupied.Occupied}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout: 10 * time.Millisecond,
		Now:     func() time.Time { return now },
		Snapshot: func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status {
			st := statuses[n]
			if n < len(statuses)-1 {
				n++
			}
			return st
		},
		SoftExit: func() { soft++ },
	})
	w.Tick() // hit1
	now = t0.Add(5 * time.Millisecond)
	w.Tick() // hit2
	now = t0.Add(15 * time.Millisecond)
	w.Tick() // hit3 + final probe occupied → no SoftExit
	if soft != 0 {
		t.Fatalf("SoftExit=%d want 0 when final occupy is occupied", soft)
	}
	if w.IdleHits() != 0 {
		t.Fatalf("hits=%d want 0 after final occupy hold", w.IdleHits())
	}
}

func TestWatchdog_notReadyHolds(t *testing.T) {
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		Ready:         func(string) bool { return false },
	})
	w.Tick()
	if w.IdleHits() != 0 {
		t.Fatalf("hits=%d want 0 when not Ready", w.IdleHits())
	}
}

func TestWatchdog_preSpaceBaselineIgnoresPostProbeFlicker(t *testing.T) {
	// Live Codex: space collapses placeholder; DEL may leave bare › briefly.
	// Probe stabilize waits until Snapshot matches pre-probe resting again.
	var soft int
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	const resting = "› Ask Codex to do anything\n  gpt-5.6-luna xhigh · ~/ws\n"
	const collapsed = "›\n  gpt-5.6-luna xhigh · ~/ws\n"
	var mu sync.Mutex
	text := resting
	postDelSnaps := 0
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout: 10 * time.Millisecond,
		Now:     func() time.Time { return now },
		Snapshot: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
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
				text = collapsed
				postDelSnaps = 0
			case "\x7f":
				text = collapsed
				postDelSnaps = 1
			}
			return nil
		},
		SoftExit: func() { soft++ },
	})
	w.Tick() // hit1; Probe restores resting before return
	now = t0.Add(5 * time.Millisecond)
	w.Tick() // hit2
	now = t0.Add(15 * time.Millisecond)
	w.Tick() // hit3 + SoftExit
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1 (placeholder restore must not reset)", soft)
	}
	if w.IdleHits() < 3 {
		t.Fatalf("hits=%d want >=3", w.IdleHits())
	}
}

func TestWatchdog_disarmedNoOp(t *testing.T) {
	var soft int
	w := New(false, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft++ },
	})
	w.Tick()
	w.Tick()
	w.Tick()
	if soft != 0 {
		t.Fatalf("SoftExit=%d want 0 when not found", soft)
	}
}

// Probe-induced viewport scroll (live Codex): after space/DEL the composer
// rematches but SnapshotText loses a leading prefix. Idle must still SoftExit —
// either Probe restores the resting snap, or condition-1 must not treat pure
// prefix scroll as activity. Today full-string Note resets hits forever.
func TestWatchdog_probeInducedViewportScrollStillSoftExits(t *testing.T) {
	const (
		prefix   = "HEAD_LINE\n  message_id=u"
		body     = "7oirumTAIL\n"
		composer = "› Ask Codex to do anything\n  footer\n"
	)
	resting := prefix + body + composer
	afterSpace := body + "›  \n  footer\n"
	afterDEL := body + composer // composer OK, head still scrolled

	var mu sync.Mutex
	text := resting
	postDelSnaps := 0
	var soft int
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout: 10 * time.Millisecond,
		Now:     func() time.Time { return now },
		Snapshot: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			// Probe stabilize: scrolled right after DEL, then full restore.
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
		SoftExit: func() { soft++ },
	})
	w.Tick() // hit1; Probe restores before return
	now = t0.Add(5 * time.Millisecond)
	w.Tick() // hit2
	now = t0.Add(15 * time.Millisecond)
	w.Tick() // SoftExit
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1 (probe must restore resting snap); hits=%d", soft, w.IdleHits())
	}
}
