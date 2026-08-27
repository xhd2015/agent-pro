package agentruncli

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/idle"
	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
)

func TestIdleWatchSchedule_shortTimeoutStartsImmediately(t *testing.T) {
	first, gap := IdleWatchSchedule(10 * time.Second)
	if first != 0 {
		t.Fatalf("first=%s want 0", first)
	}
	if gap != 5*time.Second {
		t.Fatalf("gap=%s want 5s", gap)
	}
}

func TestIdleWatchSchedule_tenMinutesWaitsThirtyThenHalfRemain(t *testing.T) {
	first, gap := IdleWatchSchedule(10 * time.Minute)
	if first != 30*time.Second {
		t.Fatalf("first=%s want 30s", first)
	}
	if gap != (10*time.Minute-30*time.Second)/2 {
		t.Fatalf("gap=%s want (10m-30s)/2", gap)
	}
}

func TestIdleWatchSchedule_zeroTimeout(t *testing.T) {
	first, gap := IdleWatchSchedule(0)
	if first != 0 || gap != 0 {
		t.Fatalf("first=%s gap=%s want 0,0", first, gap)
	}
}

func TestRunIdleWatchLoop_threeIdleSamplesThenShutdown(t *testing.T) {
	var soft, shut atomic.Int32
	w := idle.New(true, idle.Policy{ExitOnIdle: true, IdleTimeout: 10 * time.Millisecond}, idle.Watchdog{
		Timeout: 10 * time.Millisecond,
		Grace:   time.Millisecond,
		Snapshot: func() (string, error) {
			return "stable chrome", nil
		},
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft.Add(1) },
		Shutdown:      func() { shut.Add(1) },
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	idle.RunLoop(ctx, w)
	if soft.Load() != 1 {
		t.Fatalf("SoftExit=%d want 1", soft.Load())
	}
	if shut.Load() != 1 {
		t.Fatalf("Shutdown=%d want 1", shut.Load())
	}
}

func TestIdleWatchdog_snapshotChangeResetsIdleHits(t *testing.T) {
	var soft int
	snap := "v1"
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	w := idle.New(true, idle.Policy{ExitOnIdle: true}, idle.Watchdog{
		Timeout:       10 * time.Millisecond,
		Grace:         time.Second,
		Now:           func() time.Time { return now },
		Snapshot:      func() (string, error) { return snap, nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		SoftExit:      func() { soft++ },
	})
	w.Tick() // baseline + hit1
	now = t0.Add(5 * time.Millisecond)
	w.Tick() // hit 2
	snap = "v2"
	w.Tick() // change → reset
	if w.IdleHits() != 0 {
		t.Fatalf("IdleHits=%d want 0 after change", w.IdleHits())
	}
	snap = "v2"
	now = t0.Add(10 * time.Millisecond)
	w.Tick() // hit1
	now = t0.Add(15 * time.Millisecond)
	w.Tick() // hit2
	now = t0.Add(25 * time.Millisecond)
	w.Tick() // hit3 + Timeout since idleSince → SoftExit
	if soft != 1 {
		t.Fatalf("SoftExit=%d want 1 after stable empty", soft)
	}
}

func TestSleepCtx_canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if idle.SleepCtx(ctx, time.Second) {
		t.Fatal("SleepCtx on canceled ctx should be false")
	}
}
