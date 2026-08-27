package idle

import (
	"context"
	"time"
)

// SleepCtx sleeps d or returns false if ctx is done. d<=0 is a no-op success.
func SleepCtx(ctx context.Context, d time.Duration) bool {
	if ctx.Err() != nil {
		return false
	}
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// RunLoop sleeps on Schedule(Timeout) between Ticks until SoftExit+grace
// ForceShutdown, or ctx cancel. On idleHits==0 after a Tick, restarts the
// schedule from the first delay (start over).
func RunLoop(ctx context.Context, w *Watchdog) {
	if w == nil {
		return
	}
	for ctx.Err() == nil {
		first, gap := Schedule(w.Timeout)
		if !SleepCtx(ctx, first) {
			return
		}
		w.Tick()
		if w.IdleHits() == 0 {
			continue
		}
		if !SleepCtx(ctx, gap) {
			return
		}
		w.Tick()
		if w.IdleHits() == 0 {
			continue
		}
		if !SleepCtx(ctx, gap) {
			return
		}
		w.Tick()
		if !w.SoftDone() {
			continue
		}
		if !SleepCtx(ctx, w.Grace) {
			return
		}
		w.ForceShutdown()
		return
	}
}
