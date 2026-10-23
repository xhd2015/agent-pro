package sessions

import (
	"encoding/json"
	"strings"
)

// ActivityKind is the live phase of a DSH session, derived from the durable
// log plus process liveness. Values match the orchestrator contract.
type ActivityKind string

const (
	ActivityIdle        ActivityKind = "idle"
	ActivityThinking    ActivityKind = "thinking"
	ActivityResponding  ActivityKind = "responding"
	ActivityToolCalling ActivityKind = "tool_calling"
	ActivityError       ActivityKind = "error"
)

// Activity is one session's current phase. Live is process liveness, not
// stored in the log.
type Activity struct {
	Kind   ActivityKind `json:"kind"`
	Detail string       `json:"detail,omitempty"`
	Live   bool         `json:"live"`
}

// Classify derives activity from a durable event stream and whether the
// session currently holds a write lease. An open turn whose process is gone
// is an error, not idle.
func Classify(events []Event, live bool) Activity {
	return classify(events, live, true)
}

// ClassifyLog derives activity from the durable log only and never applies
// the open-turn-not-live error. Use it when process liveness could not be
// probed.
func ClassifyLog(events []Event) Activity {
	return classify(events, false, false)
}

func classify(events []Event, live, liveKnown bool) Activity {
	kind := ActivityIdle
	detail := ""
	turnOpen := false
	pendingUser := false
	for _, event := range events {
		switch event.Type {
		case "turn/start":
			turnOpen = true
			pendingUser = false
			kind = ActivityThinking
			detail = ""
		case "user/message":
			if !turnOpen {
				pendingUser = true
			}
		case "assistant/attempt":
			turnOpen = true
			pendingUser = false
			if assistantHasText(event) {
				kind = ActivityResponding
			} else {
				kind = ActivityThinking
			}
			detail = ""
		case "assistant/message":
			turnOpen = true
			pendingUser = false
			if assistantHasText(event) {
				kind = ActivityResponding
			} else {
				kind = ActivityThinking
			}
			detail = ""
		case "tool/call", "command/run":
			turnOpen = true
			pendingUser = false
			kind = ActivityToolCalling
			detail = toolName(event)
		case "tool/result", "command/done":
			if turnOpen {
				kind = ActivityThinking
				detail = ""
			}
		case "turn/end":
			turnOpen = false
			pendingUser = false
			endKind := turnEndKind(event)
			if endKind != "" && endKind != "completed" {
				kind = ActivityError
				detail = endKind
			} else {
				kind = ActivityIdle
				detail = ""
			}
		}
	}
	if pendingUser && !turnOpen {
		if live {
			kind = ActivityThinking
			detail = ""
		} else {
			kind = ActivityIdle
			detail = ""
		}
	}
	if liveKnown && turnOpen && !live {
		kind = ActivityError
		if detail == "" {
			detail = "turn open but session is not live"
		}
	}
	return Activity{Kind: kind, Detail: detail, Live: live}
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
