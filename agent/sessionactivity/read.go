package sessionactivity

import (
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func dshLookup(sessionID string, rs RunStatus) Lookup {
	return pickLookup(sessionID, rs)
}

// readDSHStatus reports one DSH Session. Kind comes from session/status first:
// running classifies the tail; false is idle; a successful map without this id
// is not_found; an unanswered RPC is error. The log is never a fallback.
func readDSHStatus(sessionID string, rs RunStatus) (Snapshot, error) {
	look := dshLookup(sessionID, rs)
	cur, err := dshTail(sessionID)
	if look.Answered && look.Present && look.Running {
		if err != nil {
			return Snapshot{}, err
		}
		snap := snapshotFromTail(cur, look)
		snap.Runner = RunnerDSH
		snap.SessionID = sessionID
		return snap, nil
	}
	snap := Snapshot{Runner: RunnerDSH, SessionID: sessionID}
	if err == nil {
		snap = snapshotFromTail(cur, look)
	} else if !look.Answered {
		snap.Kind = ErrorKind
		snap.Detail = "dsh status unreachable"
	} else if !look.Present {
		snap.Kind = NotFound
	} else {
		snap.Kind = Idle
	}
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
