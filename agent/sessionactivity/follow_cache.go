package sessionactivity

import (
	"os"
	"sync"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

type dshSession struct {
	mu   sync.Mutex
	tail *dshsessions.Tail
}

var dshSessions sync.Map

// Seams for tests: the readers must prove which of the two depths they took.
// A tail is read when the liveness answer lists the session or an outside writer
// still holds its lease; a quiescent session is answered from its header and the
// folded turn state, because reading its tail would only re-derive what the fold
// already consumed.
var (
	dshTailFn   = dshTail
	dshHeaderFn = dshHeader
)

// dshHeaderInfo is one session log's header plus the log's last write time.
type dshHeaderInfo struct {
	Summary dshsessions.Summary
	ModTime int64
}

// dshHeader reads one session log's header: identity, cwd and title, without
// decoding a single event. ModTime is the log file's last write, the cheapest
// honest stand-in for "when this session last did something": the header's own
// updated_at is the creation time, not the last activity.
func dshHeader(sessionID string) (dshHeaderInfo, error) {
	store, err := dshsessions.NewStore(dshsessions.Options{})
	if err != nil {
		return dshHeaderInfo{}, err
	}
	sum, err := store.Header(sessionID)
	if err != nil {
		return dshHeaderInfo{}, err
	}
	info := dshHeaderInfo{Summary: sum}
	if sum.Path != "" {
		if fi, statErr := os.Stat(sum.Path); statErr == nil {
			info.ModTime = fi.ModTime().UnixMilli()
		}
	}
	return info, nil
}

func dshTail(sessionID string) (*dshsessions.Tail, error) {
	raw, _ := dshSessions.LoadOrStore(sessionID, &dshSession{})
	g := raw.(*dshSession)
	g.mu.Lock()
	defer g.mu.Unlock()
	store, err := dshsessions.NewStore(dshsessions.Options{})
	if err != nil {
		return nil, err
	}
	tail, err := store.SyncTail(sessionID, g.tail)
	if err != nil {
		return nil, err
	}
	g.tail = tail
	return tail, nil
}

// liveFromLookup turns one session/status answer into the liveness observation a
// tail fold takes, keeping the instant it was sampled.
func liveFromLookup(look Lookup) dshsessions.Live {
	return dshsessions.Live{
		Running: look.Running,
		Known:   look.Answered && look.Present,
		At:      look.At,
	}
}

// snapshotFromTail classifies a session whose liveness is known: dshPhase only
// calls it for a session the registry lists, or for an open turn an outside
// writer is still executing. The tail carries the last turn/end either way, and
// that durable error or abort must not be flattened into idle; look.At travels
// with the answer so an open turn the answer predates is not called dead.
func snapshotFromTail(tail *dshsessions.Tail, look Lookup) Snapshot {
	snap := Snapshot{
		Runner:    RunnerDSH,
		SessionID: tail.Summary.ID,
		Workspace: tail.Summary.Cwd,
		Title:     tail.Summary.Title,
		UpdatedAt: tail.Summary.UpdatedAt,
		Outcome:   tail.Summary.Outcome,
	}
	act := tail.Activity(liveFromLookup(look))
	snap.Kind = Kind(act.Kind)
	snap.Detail = act.Detail
	snap.Live = act.Live
	snap.LiveStale = act.LiveStale
	return snap
}
