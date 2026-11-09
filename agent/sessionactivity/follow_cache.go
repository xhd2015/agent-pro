package sessionactivity

import (
	"sync"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

type dshSession struct {
	mu   sync.Mutex
	tail *dshsessions.Tail
}

var dshSessions sync.Map

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
	if !look.Running {
		snap.Kind = Idle
		return snap
	}
	act := tail.Activity(true, true)
	snap.Kind = Kind(act.Kind)
	snap.Detail = act.Detail
	snap.Live = true
	return snap
}
