package sessionactivity

import (
	"fmt"
	"strings"

	codexsessions "github.com/xhd2015/agent-pro/agent/codex/sessions"
	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
	groksessions "github.com/xhd2015/agent-pro/agent/grok/sessions"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

// Follow returns activity plus a committed transcript tail.
func Follow(runner, sessionID string, afterSeq int64, tail int) (Snapshot, []dshsessions.TranscriptRow, error) {
	return Default().Follow(runner, sessionID, afterSeq, tail)
}

// Follow returns activity plus a committed transcript tail.
func (p Probe) Follow(runner, sessionID string, afterSeq int64, tail int) (Snapshot, []dshsessions.TranscriptRow, error) {
	runner = strings.ToLower(strings.TrimSpace(runner))
	sessionID = strings.TrimSpace(sessionID)
	if runner == "" {
		return Snapshot{}, nil, fmt.Errorf("runner is required")
	}
	if sessionID == "" {
		return Snapshot{}, nil, fmt.Errorf("session id is required")
	}
	if tail < 0 {
		return Snapshot{}, nil, fmt.Errorf("tail must be non-negative")
	}
	switch runner {
	case RunnerDSH:
		return followDSH(sessionID, afterSeq, tail)
	case RunnerGrok:
		return followGrok(sessionID, afterSeq, tail)
	case RunnerCodex:
		return followCodex(sessionID, afterSeq, tail)
	default:
		return Snapshot{}, nil, fmt.Errorf("unknown runner: %s", runner)
	}
}

func followDSH(sessionID string, afterSeq int64, tail int) (Snapshot, []dshsessions.TranscriptRow, error) {
	store, err := dshsessions.NewStore(dshsessions.Options{})
	if err != nil {
		return Snapshot{}, nil, err
	}
	inspection, err := store.Inspect(sessionID)
	if err != nil {
		return Snapshot{}, nil, err
	}
	live := false
	liveKnown := true
	rows, err := dshsessions.ListLive(nil)
	if err != nil {
		liveKnown = false
	} else {
		for _, row := range rows {
			if row.ID == sessionID {
				live = true
				break
			}
		}
	}
	snap := snapshotFromDSH(inspection, live, liveKnown)
	snap.Runner = RunnerDSH
	snap.SessionID = sessionID
	return snap, dshsessions.Transcript(inspection.Events, afterSeq, tail), nil
}

func followGrok(sessionID string, afterSeq int64, tail int) (Snapshot, []dshsessions.TranscriptRow, error) {
	snap, err := readGrok(sessionID)
	if err != nil {
		return Snapshot{}, nil, err
	}
	home := agenttty.GrokHome()
	msgs, err := groksessions.Messages(home, sessionID, &groksessions.MessagesOpts{Limit: 0, LimitSet: true})
	if err != nil {
		return snap, nil, err
	}
	var chat []chatLine
	if msgs != nil {
		for _, m := range msgs.Messages {
			chat = append(chat, chatLine{Kind: m.Kind, Text: m.Text, Tool: m.Tool, Time: m.Timestamp.UnixMilli()})
		}
	}
	snap.Runner = RunnerGrok
	snap.SessionID = sessionID
	return snap, chatTranscript(chat, afterSeq, tail), nil
}

func followCodex(sessionID string, afterSeq int64, tail int) (Snapshot, []dshsessions.TranscriptRow, error) {
	snap, err := readCodex(sessionID)
	if err != nil {
		return Snapshot{}, nil, err
	}
	home := agenttty.CodexHome()
	msgs, err := codexsessions.Messages(home, sessionID, &codexsessions.MessagesOpts{Limit: 0, LimitSet: true})
	if err != nil {
		return snap, nil, err
	}
	var chat []chatLine
	if msgs != nil {
		for _, m := range msgs.Messages {
			chat = append(chat, chatLine{Kind: m.Kind, Text: m.Text, Tool: m.Tool, Time: m.Timestamp.UnixMilli()})
		}
	}
	snap.Runner = RunnerCodex
	snap.SessionID = sessionID
	return snap, chatTranscript(chat, afterSeq, tail), nil
}

type chatLine struct {
	Kind string
	Text string
	Tool string
	Time int64
}

func chatTranscript(msgs []chatLine, afterSeq int64, tail int) []dshsessions.TranscriptRow {
	all := make([]dshsessions.TranscriptRow, 0, len(msgs))
	var seq int64
	for _, m := range msgs {
		kind := ""
		switch m.Kind {
		case "user":
			kind = dshsessions.TranscriptUser
		case "response":
			kind = dshsessions.TranscriptAssistant
		case "tool":
			kind = dshsessions.TranscriptTool
		default:
			continue
		}
		seq++
		all = append(all, dshsessions.TranscriptRow{Seq: seq, Time: m.Time, Kind: kind, Text: m.Text, Tool: m.Tool})
	}
	return dshsessions.TranscriptFilter(all, afterSeq, tail)
}
