package sessionactivity

import (
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func readDSH(sessionID string) (Snapshot, error) {
	cur, err := dshTail(sessionID)
	if err != nil {
		return Snapshot{}, err
	}
	ids, liveKnown := dshLiveIDs()
	snap := snapshotFromTail(cur, liveKnown && ids[sessionID], liveKnown)
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
