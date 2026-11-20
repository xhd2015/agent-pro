package sessionactivity

import (
	"context"
	"errors"
	"testing"
	"time"
)

// waitFixture drives Wait without a real session log: the run-status hook says
// whether a process is live, the outcome hook says how the folded log reads.
func waitFixture(t *testing.T, look func(string) Lookup, out func(string) DurableOutcome) {
	t.Helper()
	restore := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = restore })
	t.Cleanup(func() { SetDefaultRunStatus(nil) })
	t.Cleanup(func() { SetDefaultOutcome(nil) })
	SetDefaultRunStatus(look)
	SetDefaultOutcome(out)
}

// TestWaitReturnsOnASettledTurn is the contract the verifier needs: one call, and
// the answer is the turn's ending in the same words the main session uses.
func TestWaitReturnsOnASettledTurn(t *testing.T) {
	waitFixture(t,
		func(string) Lookup { return Lookup{Answered: true, Present: true} },
		func(string) DurableOutcome {
			return DurableOutcome{Known: true, Current: true, Running: false, Last: "completed", Turns: 3}
		})
	snap, err := Wait(context.Background(), RunnerDSH, "s-done")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Idle {
		t.Fatalf("kind=%q, want idle", snap.Kind)
	}
}

// TestWaitKeepsWaitingWhileACallerHasNoDeadline: a turn runs as long as it runs.
// The verifier's old loop gave up after ten minutes and recorded a stop that
// nobody asked for.
func TestWaitKeepsWaitingWhileACallerHasNoDeadline(t *testing.T) {
	// The lookups drive the clock: while a session reads live the fold is not
	// consulted at all, so the end of the turn has to move the lookup too.
	var polls int
	finished := false
	waitFixture(t,
		func(string) Lookup {
			if !finished {
				polls++
				if polls >= 5 {
					finished = true
				}
			}
			return Lookup{Answered: true, Present: true, Running: !finished}
		},
		func(string) DurableOutcome {
			if !finished {
				return DurableOutcome{Known: true, Current: true, Running: true, Turns: 3}
			}
			return DurableOutcome{Known: true, Current: true, Running: false, Last: "completed", Turns: 3}
		})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := Wait(ctx, RunnerDSH, "s-slow"); err != nil {
		t.Fatalf("a slow turn must settle, not fail: %v", err)
	}
	if polls < 5 {
		t.Fatalf("only polled %d times", polls)
	}
}

// TestWaitReportsCancellationAsTheCallersError: a cancelled context means the
// user stopped the turn. Wait must hand that back rather than inventing a status,
// so the caller can say "aborted" and only then.
func TestWaitReportsCancellationAsTheCallersError(t *testing.T) {
	waitFixture(t,
		func(string) Lookup { return Lookup{Answered: true, Present: true, Running: true} },
		func(string) DurableOutcome {
			return DurableOutcome{Known: true, Current: true, Running: true, Turns: 3}
		})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	_, err := Wait(ctx, RunnerDSH, "s-stopped")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v, want context.Canceled", err)
	}
}

// TestWaitIgnoresATurnThatClosedBeforeTheBaseline: a prompt typed into an idle
// session leaves the previous turn's ending readable for a moment. Waiting on it
// would answer with the round before the one just asked for.
func TestWaitIgnoresATurnThatClosedBeforeTheBaseline(t *testing.T) {
	turns := int64(3)
	waitFixture(t,
		func(string) Lookup { return Lookup{Answered: true, Present: true} },
		func(string) DurableOutcome {
			return DurableOutcome{Known: true, Current: true, Running: false, Last: "completed", Turns: turns}
		})
	go func() {
		time.Sleep(20 * time.Millisecond)
		turns = 4
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	snap, err := WaitAfter(ctx, RunnerDSH, "s-next", 3)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Idle {
		t.Fatalf("kind=%q", snap.Kind)
	}
	if turns != 4 {
		t.Fatal("the test never advanced the turn count")
	}
}

// TestWaitDoesNotTreatAnUnansweredProbeAsAnEnding: on a box where the status
// source times out, "unknown" is the normal state of a session working fine.
func TestWaitDoesNotTreatAnUnansweredProbeAsAnEnding(t *testing.T) {
	waitFixture(t,
		func(string) Lookup { return Lookup{Answered: true, Present: true} },
		func(string) DurableOutcome { return DurableOutcome{} })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := Wait(ctx, RunnerDSH, "s-unknown"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v, want the deadline: an unknown fold is not an ending", err)
	}
}
