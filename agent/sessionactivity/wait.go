package sessionactivity

import (
	"context"
	"strings"
	"time"
)

// waitPollInterval is how often Wait re-reads a session that has not settled.
//
// A variable so a test can shorten it; production has no reason to change it.
var waitPollInterval = 500 * time.Millisecond

// Wait follows a session until its newest turn closes, and answers with the
// snapshot that closed it.
//
// It is the shared way to "run a turn and see how it ended", which is what a
// verifier does. The alternative -- each caller writing its own poll loop -- is
// how a verification grew a ten-minute deadline the main session never had, a
// status vocabulary of its own, and a stop that could not be told from a
// timeout. One implementation, so one answer.
//
// The only thing that ends the wait early is ctx: a caller cancelling it means
// the *user* stopped this turn, and the returned error is ctx.Err() so the
// caller can say "aborted" rather than inventing an outcome. There is no
// deadline here on purpose -- a turn runs as long as it runs, exactly as it does
// for the main session, which no timer interrupts.
func Wait(ctx context.Context, runner, sessionID string) (Snapshot, error) {
	return WaitAfter(ctx, runner, sessionID, -1)
}

// WaitAfter is Wait, but only accepts a turn that closed *after* the given
// count of closed turns.
//
// A prompt typed into a session that is momentarily idle leaves a window where
// the log still describes the previous turn: the turn count has not moved yet.
// Waiting on that snapshot would answer with the round before the one just
// asked for. Passing the count observed before handing over the prompt closes
// the window, and is why this is a comparison rather than a delay.
//
// A negative baseline disables the comparison, for a caller that has no earlier
// observation.
func WaitAfter(ctx context.Context, runner, sessionID string, afterTurns int64) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	for {
		snap, err := Read(runner, sessionID)
		if err != nil {
			// A probe that cannot answer is not an ending. Keep trying until the
			// context says otherwise: on a box where the status source times out,
			// "unknown" is the normal state of a session working perfectly well.
			select {
			case <-ctx.Done():
				return Snapshot{}, ctx.Err()
			case <-time.After(waitPollInterval):
			}
			continue
		}
		if settled(snap) && turnsAdvanced(snap, afterTurns) {
			return snap, nil
		}
		select {
		case <-ctx.Done():
			return Snapshot{}, ctx.Err()
		case <-time.After(waitPollInterval):
		}
	}
}

// settled reports whether a snapshot describes a turn that is over.
//
// The log is what decides, and `Kind` is its verdict: settledSnapshot maps a
// closed turn to Idle, Aborted or ErrorKind, and everything still in flight to a
// phase of its own. Reading `Outcome` instead would miss a turn that closed with
// no recorded ending -- settledSnapshot deliberately maps that to Idle too.
//
// Liveness cannot decide it: a session between two turns is not live either.
// An unanswered probe is not an ending, and a stale read is a statement about
// the past, not about the newest turn.
func settled(snap Snapshot) bool {
	if snap.Unanswered || snap.Live || snap.LiveStale {
		return false
	}
	switch snap.Kind {
	case Idle, Aborted, ErrorKind, NotFound:
		return true
	default:
		return false
	}
}

// turnsAdvanced reports whether the settled snapshot is about a turn newer than
// the baseline.
//
// The baseline is a count on the fold, not a field of Snapshot, so this asks the
// fold directly. A negative baseline accepts whatever settles first.
func turnsAdvanced(snap Snapshot, afterTurns int64) bool {
	if afterTurns < 0 {
		return true
	}
	return ClosedTurns(snap.Runner, snap.SessionID) > afterTurns
}

// ClosedTurns is how many turns a session's log has closed. It is the number a
// caller records before handing over a prompt, so WaitAfter can tell the round
// it asked for from the one before it.
func ClosedTurns(runner, sessionID string) int64 {
	runner = strings.ToLower(strings.TrimSpace(runner))
	sessionID = strings.TrimSpace(sessionID)
	if runner != RunnerDSH || sessionID == "" {
		return 0
	}
	return pickOutcome(sessionID, Default().outcomeSource()).Turns
}
