package cronspec

import (
	"context"
	"errors"
	"time"
)

// ErrNeverFires reports a valid spec that cannot produce another fire time.
var ErrNeverFires = errors.New("schedule never fires")

// RunOptions configures Run.
type RunOptions struct {
	// MaxRuns stops the loop after this many cycles. 0 means unlimited.
	MaxRuns int
	// Now is the clock; nil means time.Now.
	Now func() time.Time
	// Logf receives schedule progress lines; nil means silent.
	Logf func(format string, args ...any)
}

// Run calls fn immediately, then once per scheduled fire, until ctx is
// cancelled or MaxRuns cycles have completed. The first cycle runs without an
// initial wait, matching kool for-every. Errors returned by fn are reported
// through Logf and do not stop the loop; Run returns nil on clean cancellation.
func Run(ctx context.Context, s Schedule, opts RunOptions, fn func(context.Context) error) error {
	if s == nil {
		return errors.New("cronspec: nil schedule")
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	// prev tracks the last scheduled fire so intervals stay drift-free.
	prev := now()
	runs := 0
	for {
		if err := fn(ctx); err != nil {
			logf("cycle failed: %v", err)
		}
		runs++
		if opts.MaxRuns > 0 && runs >= opts.MaxRuns {
			return nil
		}
		if ctx.Err() != nil {
			return nil
		}

		next := s.Next(prev)
		if next.IsZero() {
			return ErrNeverFires
		}
		// Skip fires missed while a previous cycle ran.
		current := now()
		for !next.After(current) {
			prev = next
			next = s.Next(prev)
			if next.IsZero() {
				return ErrNeverFires
			}
		}
		prev = next

		logf("next run at %s", next.Format(time.RFC3339))
		timer := time.NewTimer(next.Sub(current))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil
		case <-timer.C:
		}
	}
}
