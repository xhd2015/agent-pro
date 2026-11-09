package sessionactivity

import (
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func dshLookup(sessionID string, rs RunStatus) Lookup {
	return pickLookup(sessionID, rs)
}

// readDSHStatus reports one DSH Session. session/status decides how deep the
// log is read: an unanswered RPC is error without touching the disk, an id
// missing from an answered map is only known to be not loaded (so its metadata
// comes from the log header, never a tail), and a listed session is classified
// from its tail with liveness taken from the answer, so a durable turn/end
// error or abort still shows through after the process is gone. not_found is
// reserved for a header that proves the log itself is gone.
func readDSHStatus(sessionID string, rs RunStatus) (Snapshot, error) {
	look := dshLookup(sessionID, rs)
	if !look.Answered {
		return Snapshot{
			Runner:    RunnerDSH,
			SessionID: sessionID,
			Kind:      ErrorKind,
			Detail:    "dsh status unreachable",
		}, nil
	}
	if !look.Present {
		return snapshotFromHeader(sessionID), nil
	}
	cur, err := dshTailFn(sessionID)
	if err != nil {
		if look.Running {
			return Snapshot{}, err
		}
		return snapshotFromHeader(sessionID), nil
	}
	snap := snapshotFromTail(cur, look)
	snap.Runner = RunnerDSH
	snap.SessionID = sessionID
	return snap, nil
}

func readGrok(sessionID string) (Snapshot, error) {
	home := agenttty.GrokHome()
	st, err := groksessions.Status(home, sessionID, true, nil)
	if err != nil {
		return Snapshot{}, err
	}
	live := st.State == "running"
	msgs, err := groksessions.Messages(home, sessionID, &groksessions.MessagesOpts{Limit: 1, LimitSet: true})
	if err != nil {
		return Snapshot{}, err
	}
	lastKind, tool := "", ""
	if msgs != nil && len(msgs.Messages) > 0 {
		last := msgs.Messages[len(msgs.Messages)-1]
		lastKind, tool = last.Kind, last.Tool
	}
	kind, detail := ClassifyChat(lastKind, tool, live)
	snap := Snapshot{
		Runner:    RunnerGrok,
		SessionID: sessionID,
		Kind:      kind,
		Detail:    detail,
		Live:      live,
	}
	if info, err := groksessions.Info(home, sessionID); err == nil && info != nil {
		snap.Workspace = info.CWD
		snap.Title = info.Title
		if !info.UpdatedAt.IsZero() {
			snap.UpdatedAt = info.UpdatedAt.UnixMilli()
		}
	}
	return snap, nil
}

func readCodex(sessionID string) (Snapshot, error) {
	home := agenttty.CodexHome()
	st, err := codexsessions.Status(home, sessionID, true, nil)
	if err != nil {
		return Snapshot{}, err
	}
	live := st.State == "running"
	msgs, err := codexsessions.Messages(home, sessionID, &codexsessions.MessagesOpts{Limit: 1, LimitSet: true})
	if err != nil {
		return Snapshot{}, err
	}
	lastKind, tool := "", ""
	if msgs != nil && len(msgs.Messages) > 0 {
		last := msgs.Messages[len(msgs.Messages)-1]
		lastKind, tool = last.Kind, last.Tool
	}
	kind, detail := ClassifyChat(lastKind, tool, live)
	snap := Snapshot{
		Runner:    RunnerCodex,
		SessionID: sessionID,
		Kind:      kind,
		Detail:    detail,
		Live:      live,
	}
	if info, err := codexsessions.Info(home, sessionID, 0); err == nil && info != nil {
		snap.Workspace = info.CWD
		snap.Title = info.Title
		if !info.StartedAt.IsZero() {
			snap.UpdatedAt = info.StartedAt.UnixMilli()
		}
	}
	return snap, nil
}
