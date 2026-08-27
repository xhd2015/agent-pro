package idle

import (
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

func TestWatchdog_unknownResets(t *testing.T) {
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "stable", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Unknown },
	})
	w.Tick()
	if w.IdleHits() != 0 {
		t.Fatalf("hits=%d want 0 on unknown", w.IdleHits())
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
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
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
	w.Tick() // hit1, probe empty
	w.Tick() // hit2, probe empty
	w.Tick() // hit3, final probe occupied → no SoftExit
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

func TestWatchdog_probeSideEffectDoesNotReset(t *testing.T) {
	// Simulate occupy probe mutating the resting snap between ticks; after
	// re-baseline the next identical post-probe snap must still count as idle.
	var soft int
	n := 0
	snaps := []string{"stable", "stable+probe", "stable+probe", "stable+probe", "stable+probe"}
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot: func() (string, error) {
			s := snaps[n]
			if n < len(snaps)-1 {
				n++
			}
			return s, nil
		},
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft++ },
	})
	// Tick1: baseline "stable", probe, re-baseline "stable+probe", hit1
	w.Tick()
	// Tick2/3: resting "stable+probe" unchanged vs re-baseline → hits 2,3
	w.Tick()
	w.Tick()
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1 (probe residue must not poison changed)", soft)
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
