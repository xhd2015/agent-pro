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

// Seams for tests: the readers must prove which of the two paths they took.
// A session the answered map lists owns its phase in the log; a session the map
// does not list must stop at the header.
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

// snapshotFromTail answers a session session/status lists. Its liveness is then
// known either way, so the tail is classified with look.Running as the live
// flag: a not-running session still carries its last turn/end in the log, and
// that durable error or abort must not be flattened into idle.
func snapshotFromTail(tail *dshsessions.Tail, look Lookup) Snapshot {
	snap := Snapshot{
		Runner:    RunnerDSH,
		SessionID: tail.Summary.ID,
		Workspace: tail.Summary.Cwd,
		Title:     tail.Summary.Title,
		UpdatedAt: tail.Summary.UpdatedAt,
		Outcome:   tail.Summary.Outcome,
	}
	if !look.Answered {
		snap.Kind = ErrorKind
		snap.Detail = "dsh status unreachable"
		return snap
	}
	if !look.Present {
		snap.Kind = NotFound
		return snap
	}
	// The session is present, so liveness is known: classify the tail either way.
	// A not-running session still carries its last turn/end in the log, and that
	// durable error or abort must not be flattened into idle.
	act := tail.Activity(look.Running, true)
	snap.Kind = Kind(act.Kind)
	snap.Detail = act.Detail
	snap.Live = act.Live
	return snap
}

// snapshotFromHeader answers a session the answered map does not list: the
// daemon has no state for it, so nothing past the header is read. not_found
// therefore means the header lookup proved there is no log at all — an id the
// daemon merely never loaded is idle, not missing.
func snapshotFromHeader(sessionID string) Snapshot {
	snap := Snapshot{Runner: RunnerDSH, SessionID: sessionID}
	info, err := dshHeaderFn(sessionID)
	if err != nil {
		snap.Kind = NotFound
		snap.Detail = err.Error()
		return snap
	}
	if info.Summary.ID != "" {
		snap.SessionID = info.Summary.ID
	}
	snap.Kind = Idle
	snap.Workspace = info.Summary.Cwd
	snap.Title = info.Summary.Title
	snap.UpdatedAt = info.ModTime
	if snap.UpdatedAt == 0 {
		snap.UpdatedAt = info.Summary.UpdatedAt
	}
	return snap
}
