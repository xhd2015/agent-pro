package sessionactivity

import (
	"github.com/xhd2015/agent-pro/agent/dsh/consumers"
	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// Details the DSH readers report when a turn did not end because the provider
// answered. They are constants because callers must tell such a turn from a
// provider failure: auto-retry has nothing to re-send for a turn nobody
// answered.
const (
	// DetailInterrupted marks a turn the log closed as interrupted, or an open
	// turn whose writer is gone.
	DetailInterrupted = consumers.OutcomeInterrupted
	// DetailLogAhead marks an open turn newer than the status answer: the answer
	// could not testify about it. It is a placeholder, not a phase.
	DetailLogAhead = "turn state is newer than the status answer"
	// DetailWriterUnknown marks an open turn whose write lease could not be
	// observed. It is a placeholder, not a phase.
	DetailWriterUnknown = "session writer is not known"
	// DetailNotFolded marks a session whose folded turn state is not usable yet.
	// It is a placeholder, not a phase.
	DetailNotFolded = "session outcome not folded yet"
	// DetailUnreadable marks a log the readers could not read at all. It is a
	// placeholder, not a phase.
	DetailUnreadable = "session log is unreadable"
	// DetailTurnOpen marks an error whose turn never ended because the process
	// that ran it is gone. It is shared with the tail classifier so both depths
	// report the death identically.
	DetailTurnOpen = dshsessions.DetailTurnOpen
)

// IsLivenessDetail reports whether a Detail describes the turn's liveness rather
// than a failure the provider reported. Such a turn got no answer, so retry
// automation has nothing to re-send and no rule should match it.
func IsLivenessDetail(detail string) bool {
	switch detail {
	case DetailTurnOpen, DetailInterrupted, DetailLogAhead, DetailWriterUnknown, DetailNotFolded, DetailUnreadable:
		return true
	}
	return false
}

// dshSources are the three independent observations a DSH verdict is composed
// from, each injectable so a test can prove which one decided.
type dshSources struct {
	// runStatus answers "is a turn running in the agent registry right now?".
	runStatus RunStatus
	// outcome answers "did the newest turn in the log close?".
	outcome DurableOutcomeSource
	// writers answers "is a process still holding the log's write lease?".
	writers WriterSource
}

func dshSourcesOf(p Probe) dshSources {
	return dshSources{
		runStatus: p.runStatusHook(),
		outcome:   p.outcomeSource(),
		writers:   p.writerSource(),
	}
}

// dshPhase composes one DSH session's phase from the three observations. Each
// answers a question the others cannot, and each is allowed to be silent —
// silence is never flattened into idle, because that is what reported a session
// killed with its turn open as finished.
//
// The registry wins when it lists a running turn. When it does not, the folded
// log decides whether a turn is still open; only then does the write lease
// matter, and only to tell a turn executing outside the registry from one whose
// writer is gone.
func dshPhase(sessionID string, src dshSources, look Lookup) (Snapshot, error) {
	if !look.Answered {
		return Snapshot{
			Runner:     RunnerDSH,
			SessionID:  sessionID,
			Kind:       ErrorKind,
			Detail:     "dsh status unreachable",
			Unanswered: true,
		}, nil
	}
	if look.Running {
		return dshTailSnapshot(sessionID, look)
	}
	out := pickOutcome(sessionID, src.outcome)
	if out.Known && out.Current {
		if out.Running {
			return openTurnSnapshot(sessionID, src, look, out), nil
		}
		return settledSnapshot(sessionID, out), nil
	}
	// No usable fold: the tail still decides a session the daemon lists, and for
	// one it does not list there is nothing left to decide with, so abstain
	// rather than inherit "no daemon entry means idle".
	if look.Present {
		return dshTailSnapshot(sessionID, look)
	}
	return pendingSnapshot(sessionID, DetailNotFolded), nil
}

// dshTailSnapshot classifies one session's tail with the given liveness.
func dshTailSnapshot(sessionID string, look Lookup) (Snapshot, error) {
	cur, err := dshTailFn(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	snap := snapshotFromTail(cur, look)
	snap.Runner = RunnerDSH
	snap.SessionID = sessionID
	return snap, nil
}

// openTurnSnapshot answers a session whose fold left a turn open while the
// registry reported no running turn: either a process outside the registry is
// executing it, or the writer died with the turn open.
func openTurnSnapshot(sessionID string, src dshSources, look Lookup, out DurableOutcome) Snapshot {
	// An answer sampled before the newest write cannot testify about the turn
	// that write opened, so neither can a lease sample taken now.
	if !look.At.IsZero() && out.Modified.After(look.At) {
		return pendingSnapshot(sessionID, DetailLogAhead)
	}
	held, known := pickWriter(sessionID, src.writers)
	if !known {
		return pendingSnapshot(sessionID, DetailWriterUnknown)
	}
	if held {
		// A live writer outside the registry: the turn runs, and the tail still
		// carries its phase and tool name.
		running := Lookup{Answered: true, Present: true, Running: true, At: look.At}
		if snap, err := dshTailSnapshot(sessionID, running); err == nil {
			return snap
		}
		return pendingSnapshot(sessionID, DetailUnreadable)
	}
	// Nobody is writing a turn that never ended: the process that ran it is gone
	// and the turn will never close. That is a killed turn, not idle.
	snap := Snapshot{
		Runner:    RunnerDSH,
		SessionID: sessionID,
		Kind:      ErrorKind,
		Detail:    DetailInterrupted,
		Outcome:   consumers.OutcomeInterrupted,
	}
	return withHeader(snap, sessionID)
}

// settledSnapshot answers a session whose log is quiescent and whose newest
// turn closed: the tail would only re-derive what the last turn/end already
// says. The mapping mirrors the tail classifier, so a turn closed as
// interrupted or aborted still reads as a failure after the process is gone.
func settledSnapshot(sessionID string, out DurableOutcome) Snapshot {
	snap := Snapshot{Runner: RunnerDSH, SessionID: sessionID, Outcome: out.Last}
	switch out.Last {
	case "", consumers.OutcomeCompleted:
		snap.Kind = Idle
	case consumers.OutcomeAborted:
		snap.Kind = Aborted
		snap.Detail = out.Last
	default:
		snap.Kind = ErrorKind
		snap.Detail = out.Last
	}
	return withHeader(snap, sessionID)
}

// pendingSnapshot answers a session whose phase could not be decided: the
// metadata comes from the header, and the kind is an unknown placeholder so no
// surface reports idle, finished or failed. A header that proves the log itself
// is gone stays not_found.
func pendingSnapshot(sessionID, detail string) Snapshot {
	info, err := dshHeaderFn(sessionID)
	if err != nil {
		return Snapshot{Runner: RunnerDSH, SessionID: sessionID, Kind: NotFound, Detail: err.Error()}
	}
	snap := Snapshot{
		Runner:     RunnerDSH,
		SessionID:  sessionID,
		Kind:       ErrorKind,
		Detail:     detail,
		Unanswered: true,
	}
	return withHeaderInfo(snap, info)
}

// withHeader fills a decided snapshot's metadata from the log header, which
// costs one stat. A header it cannot read leaves the decision intact: the fold
// just read that log, so a vanished header is not evidence about the turn.
func withHeader(snap Snapshot, sessionID string) Snapshot {
	if snap.SessionID == "" {
		snap.SessionID = sessionID
	}
	info, err := dshHeaderFn(sessionID)
	if err != nil {
		return snap
	}
	return withHeaderInfo(snap, info)
}

func withHeaderInfo(snap Snapshot, info dshHeaderInfo) Snapshot {
	if info.Summary.ID != "" {
		snap.SessionID = info.Summary.ID
	}
	snap.Workspace = info.Summary.Cwd
	snap.Title = info.Summary.Title
	snap.UpdatedAt = info.ModTime
	if snap.UpdatedAt == 0 {
		snap.UpdatedAt = info.Summary.UpdatedAt
	}
	return snap
}
