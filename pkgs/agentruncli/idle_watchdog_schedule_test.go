package agentruncli

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
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
	w := NewIdleWatchdog(true, IdlePolicy{ExitOnIdle: true, IdleTimeout: 10 * time.Millisecond}, IdleWatchdog{
		Timeout: 10 * time.Millisecond,
		Grace:   time.Millisecond,
		Sample: func() IdleSample {
			return IdleSample{Sendable: true, Screen: "idle", InputBox: "empty"}
		},
		SoftExit: func() { soft.Add(1) },
		Shutdown: func() { shut.Add(1) },
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	runIdleWatchLoop(ctx, w)
	if soft.Load() != 1 {
		t.Fatalf("SoftExit=%d want 1", soft.Load())
	}
	if shut.Load() != 1 {
		t.Fatalf("Shutdown=%d want 1", shut.Load())
	}
}

func TestSleepCtx_canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if sleepCtx(ctx, time.Second) {
		t.Fatal("sleepCtx on canceled ctx should be false")
	}
}
