package sessionactivity

import (
	"sync"
	"time"
)

// DurableOutcome is what a session's own log says about its turns, folded by a
// consumer bus instead of read from the tail.
//
// Liveness and the log are sampled independently and either one can be silent.
// session/status knows only the sessions a daemon has loaded, so for any other
// session it cannot say whether a turn is running — and a log that stops looks
// the same whether the turn is still executing, the process was killed, or the
// daemon started elsewhere. A fold closes that gap because it consumed every row
// up to its own Size: it knows whether the newest turn/start still has no
// turn/end, and the writer lease says whether anyone is still writing that log.
// An open turn nobody writes is a turn that will never close.
type DurableOutcome struct {
	// Known is false before the fold has read the session at all. It is not
	// "idle": no fold means the turn state was never observed.
	Known bool
	// Current is whether the fold consumed the whole log as of Modified. A fold
	// that is behind cannot testify about the newest turn, so its Running is a
	// statement about the past, not about now.
	Current bool
	// Running is whether the newest turn/start has no matching turn/end yet.
	Running bool
	// Last is the end kind of the last closed turn: completed, aborted,
	// interrupted, max-tokens, error. Empty when no turn has closed.
	Last string
	// Turns is how many turns closed. It makes "the fold saw the end of the
	// turn the answer was about" a comparison rather than a guess.
	Turns int64
	// Modified is the log's last write, sampled with the fold. The readers
	// order it against Lookup.At: an answer sampled before the newest write
	// cannot testify about the turn that write opened.
	Modified time.Time
}

// DurableOutcomeSource answers one session's folded turn state. A zero value
// means the source has no usable fold for the session, never that the session
// is idle.
type DurableOutcomeSource func(sessionID string) DurableOutcome

// WriterSource reports whether a live process still holds one session's write
// lease. Held is the answer; Known says whether the source could answer at all.
// A process scan can fail, and an unknown lease must never be read as a
// released one: that would turn an unreadable process table into "every open
// turn is dead".
type WriterSource func(sessionID string) (held bool, known bool)

var (
	durableMu sync.RWMutex
	durableFn DurableOutcomeSource

	writerMu sync.RWMutex
	writerFn WriterSource
)

// SetDefaultOutcome installs the process-wide DurableOutcomeSource that the
// default DSH reader consults; a nil hook uninstalls it and leaves the folded
// turn state unknown, which makes the reader abstain rather than guess.
//
// Only the DSH reader consults it. Grok and Codex report liveness from their own
// registries, so installing a hook can never change their answer.
func SetDefaultOutcome(fn DurableOutcomeSource) {
	durableMu.Lock()
	durableFn = fn
	durableMu.Unlock()
}

// defaultOutcome returns the installed hook, or nil when none is installed.
func defaultOutcome() DurableOutcomeSource {
	durableMu.RLock()
	defer durableMu.RUnlock()
	return durableFn
}

// SetDefaultWriter installs the process-wide WriterSource that tells the DSH
// reader whether a session's write lease is still held; a nil hook uninstalls
// it, and the reader then abstains from calling an open turn dead.
func SetDefaultWriter(fn WriterSource) {
	writerMu.Lock()
	writerFn = fn
	writerMu.Unlock()
}

// defaultWriter returns the installed hook, or nil when none is installed.
func defaultWriter() WriterSource {
	writerMu.RLock()
	defer writerMu.RUnlock()
	return writerFn
}

// outcomeSource is the hook this probe consults: its own field when set, else
// the installed default.
func (p Probe) outcomeSource() DurableOutcomeSource {
	if p.Outcome != nil {
		return p.Outcome
	}
	return defaultOutcome()
}

// writerSource is the hook this probe consults for the write lease: its own
// field when set, else the installed default.
func (p Probe) writerSource() WriterSource {
	if p.Writers != nil {
		return p.Writers
	}
	return defaultWriter()
}

func pickOutcome(sessionID string, fn DurableOutcomeSource) DurableOutcome {
	if fn == nil {
		return DurableOutcome{}
	}
	return fn(sessionID)
}

func pickWriter(sessionID string, fn WriterSource) (bool, bool) {
	if fn == nil {
		return false, false
	}
	return fn(sessionID)
}
