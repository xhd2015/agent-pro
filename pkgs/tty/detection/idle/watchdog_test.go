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
	// Live Codex: space probe collapses "Ask Codex to do anything" to bare ›
	// (DEL does not restore the hint). Codex redraws the hint before the next
	// Tick. Resting compare must keep the pre-space baseline so that restore
	// does not look like session activity.
	var soft int
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	const resting = "› Ask Codex to do anything\n  gpt-5.6-luna xhigh · ~/ws\n"
	const collapsed = "›\n  gpt-5.6-luna xhigh · ~/ws\n"
	var mu sync.Mutex
	text := resting
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Timeout: 10 * time.Millisecond,
		Now:     func() time.Time { return now },
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
				text = collapsed
			case "\x7f":
				// Live Codex: DEL leaves bare ›; hint returns later.
				text = collapsed
			}
			return nil
		},
		SoftExit: func() { soft++ },
	})
	w.Tick() // hit1; baseline = resting (not collapsed)
	mu.Lock()
	text = resting // Codex redraws placeholder between ticks
	mu.Unlock()
	now = t0.Add(5 * time.Millisecond)
	w.Tick() // resting vs resting → hit2
	mu.Lock()
	text = resting
	mu.Unlock()
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
