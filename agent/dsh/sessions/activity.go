package sessions

import (
	"encoding/json"
	"strings"
	"time"
)

// ActivityKind is the live phase of a DSH session, derived from the durable
// log plus process liveness. Values match the orchestrator contract.
type ActivityKind string

const (
	ActivityIdle        ActivityKind = "idle"
	ActivityThinking    ActivityKind = "thinking"
	ActivityResponding  ActivityKind = "responding"
	ActivityToolCalling ActivityKind = "tool_calling"
	ActivityWaiting     ActivityKind = "waiting"
	ActivityError       ActivityKind = "error"
	ActivityAborted     ActivityKind = "aborted"
)

// DetailTurnOpen is the detail of an error whose turn never ended because the
// process that ran it is gone. It is not a provider failure: nobody answered the
// turn, so a retry has nothing to re-send.
const DetailTurnOpen = "turn open but session is not live"

// Live is one process-liveness observation of a session: whether it runs, and
// when that was observed.
//
// At is the instant the answer was sampled: the moment the request was issued,
// which is never later than the reply. It exists because liveness is answered by
// a cached batch RPC while the phase is folded from the log: without the instant
// the two facts cannot be ordered, and a turn started in between is reported as
// a dead session.
type Live struct {
	Running bool
	Known   bool
	At      time.Time
}

// Activity is one session's current phase. Live is process liveness, not
// stored in the log.
type Activity struct {
	Kind   ActivityKind `json:"kind"`
	Detail string       `json:"detail,omitempty"`
	Live   bool         `json:"live"`
	// LiveStale marks a phase folded from a log whose open turn is newer than
	// the liveness answer. Live is then false without the turn having failed:
	// the answer could not testify about it.
	LiveStale bool `json:"live_stale,omitempty"`
}

// Classify derives activity from a durable event stream and one liveness
// observation of the same session. An open turn whose process is gone is an
// error, not idle — but only an answer at least as new as that turn's start can
// say so, which is why Live carries its instant.
func Classify(events []Event, live Live) Activity {
	return classify(events, live)
}

// ClassifyLog derives activity from the durable log only and never applies
// the open-turn-not-live error. Use it when process liveness could not be
// probed.
func ClassifyLog(events []Event) Activity {
	return classify(events, Live{})
}

type classifyState struct {
	kind        ActivityKind
	detail      string
	turnOpen    bool
	pendingUser bool
	// openedAt is when the open turn started, in the log's epoch ms, taken from
	// the earliest event that opened it. Zero means the log does not say.
	openedAt int64
}

func classify(events []Event, live Live) Activity {
	var state classifyState
	for _, event := range events {
		state.add(event)
	}
	return state.finish(live)
}

func (s *classifyState) add(event Event) {
	switch event.Type {
	case "turn/start":
		s.open(event.Time)
		s.pendingUser = false
		s.kind = ActivityThinking
		s.detail = ""
	case "user/message":
		if !s.turnOpen {
			s.pendingUser = true
		}
	case "assistant/attempt":
		s.open(event.Time)
		s.pendingUser = false
		if assistantHasText(event) {
			s.kind = ActivityResponding
		} else {
			s.kind = ActivityThinking
		}
		s.detail = ""
	case "assistant/message":
		s.open(event.Time)
		s.pendingUser = false
		if assistantHasText(event) {
			s.kind = ActivityResponding
		} else {
			s.kind = ActivityThinking
		}
		s.detail = ""
	case "tool/call", "command/run":
		s.open(event.Time)
		s.pendingUser = false
		name := toolName(event)
		if name == "ask_user_question" {
			s.kind = ActivityWaiting
		} else {
			s.kind = ActivityToolCalling
		}
		s.detail = name
	case "tool/result", "command/done":
		if s.turnOpen {
			s.kind = ActivityThinking
			s.detail = ""
		}
	case "turn/end":
		s.turnOpen = false
		s.pendingUser = false
		s.openedAt = 0
		endKind := turnEndKind(event)
		if endKind == "aborted" {
			s.kind = ActivityAborted
			s.detail = endKind
		} else if endKind != "" && endKind != "completed" {
			s.kind = ActivityError
			s.detail = endKind
		} else {
			s.kind = ActivityIdle
			s.detail = ""
		}
	}
}

// open records that a turn is open at the log instant at, keeping the earliest
// opening event of the current turn. A fold that started mid-turn can only place
// the start later than the real one, which makes an older liveness answer abstain
// more often rather than blame the session.
func (s *classifyState) open(at int64) {
	if !s.turnOpen || s.openedAt == 0 {
		s.openedAt = at
	}
	s.turnOpen = true
}

// finish turns the folded log plus one liveness observation into the session's
// phase.
//
// The two inputs are sampled at different instants: the log is as fresh as this
// fold, the liveness answer can be a cached batch. A not-running answer therefore
// only means "this open turn is dead" when it was sampled at or after the turn
// started; an older answer cannot testify about this turn, and the log alone then
// says what the session is doing. Suppressing a false alarm is the only possible
// error here: an answer sampled after the start is never discarded.
func (s classifyState) finish(live Live) Activity {
	kind, detail := s.kind, s.detail
	if kind == "" {
		kind = ActivityIdle
	}
	if s.pendingUser && !s.turnOpen {
		if live.Running {
			kind = ActivityThinking
			detail = ""
		} else {
			kind = ActivityIdle
			detail = ""
		}
	}
	if live.Known && s.turnOpen && !live.Running {
		if s.livenessAnsweredTurn(live.At) {
			kind = ActivityError
			if detail == "" {
				detail = DetailTurnOpen
			}
			return Activity{Kind: kind, Detail: detail, Live: live.Running}
		}
		return Activity{Kind: kind, Detail: detail, Live: live.Running, LiveStale: true}
	}
	return Activity{Kind: kind, Detail: detail, Live: live.Running}
}

// livenessAnsweredTurn reports whether an answer sampled at at can testify about
// the open turn. An unrecorded instant on either side keeps the historical
// verdict: the log and the answer came from the same read, as before Live existed.
func (s classifyState) livenessAnsweredTurn(at time.Time) bool {
	if s.openedAt == 0 || at.IsZero() {
		return true
	}
	return at.UnixMilli() >= s.openedAt
}

func assistantHasText(event Event) bool {
	content := messageContent(event.Data)
	for _, block := range content {
		if strings.TrimSpace(block["type"]) != "text" {
			continue
		}
		if strings.TrimSpace(block["text"]) != "" {
			return true
		}
	}
	return false
}

func messageContent(raw json.RawMessage) []map[string]string {
	var data map[string]json.RawMessage
	if json.Unmarshal(raw, &data) != nil || data == nil {
		return nil
	}
	body := data["content"]
	if msg, ok := data["message"]; ok {
		var inner map[string]json.RawMessage
		if json.Unmarshal(msg, &inner) == nil && inner != nil {
			if _, ok := inner["content"]; ok {
				body = inner["content"]
			}
		}
	}
	var blocks []map[string]json.RawMessage
	if json.Unmarshal(body, &blocks) != nil {
		return nil
	}
	out := make([]map[string]string, 0, len(blocks))
	for _, block := range blocks {
		item := map[string]string{}
		for _, key := range []string{"type", "text"} {
			var value string
			if json.Unmarshal(block[key], &value) == nil {
				item[key] = value
			}
		}
		out = append(out, item)
	}
	return out
}

func toolName(event Event) string {
	var data map[string]json.RawMessage
	if json.Unmarshal(event.Data, &data) != nil || data == nil {
		return ""
	}
	for _, key := range []string{"name", "tool", "command"} {
		var value string
		if json.Unmarshal(data[key], &value) == nil {
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func turnEndKind(event Event) string {
	var data map[string]json.RawMessage
	if json.Unmarshal(event.Data, &data) != nil || data == nil {
		return ""
	}
	var reason map[string]json.RawMessage
	if json.Unmarshal(data["reason"], &reason) != nil || reason == nil {
		return ""
	}
	var kind string
	if json.Unmarshal(reason["kind"], &kind) != nil {
		return ""
	}
	return strings.TrimSpace(kind)
}
