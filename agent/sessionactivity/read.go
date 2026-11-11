package sessionactivity

import (
	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func dshLookup(sessionID string, rs RunStatus) Lookup {
	return pickLookup(sessionID, rs)
}

// readDSHStatus reports one DSH Session, composing the verdict in dshPhase: the
// agent registry says whether a turn runs, the folded log says whether the
// newest turn ever closed, and the write lease says whether an open turn still
// has a writer. An id the registry does not list is no longer read as idle; its
// log decides, so a turn killed with the process is an error rather than a
// finished session. not_found is reserved for a header that proves the log
// itself is gone.
func readDSHStatus(sessionID string, src dshSources) (Snapshot, error) {
	return dshPhase(sessionID, src, dshLookup(sessionID, src.runStatus))
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
