package sessions

import (
	"encoding/json"
	"strings"
)

const (
	TranscriptUser      = "user"
	TranscriptAssistant = "assistant"
	TranscriptTool      = "tool"
)

// TranscriptRow is one display line of a live session log.
type TranscriptRow struct {
	Seq  int64  `json:"seq"`
	Time int64  `json:"time"`
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
	Tool string `json:"tool,omitempty"`
}

// Transcript projects committed user, assistant, and tool/call events.
// afterSeq > 0 returns only later seqs (no tail trim). afterSeq == 0 keeps
// the last tail rows (tail <= 0 means 24).
func Transcript(events []Event, afterSeq int64, tail int) []TranscriptRow {
	out := make([]TranscriptRow, 0)
	for _, event := range events {
		row, ok := transcriptRow(event)
		if !ok {
			continue
		}
		out = append(out, row)
	}
	return TranscriptFilter(out, afterSeq, tail)
}

// TranscriptFilter applies afterSeq (exclusive) and tail. afterSeq > 0
// returns every later row; otherwise the last tail rows remain (default 24).
func TranscriptFilter(rows []TranscriptRow, afterSeq int64, tail int) []TranscriptRow {
	if tail <= 0 {
		tail = 24
	}
	if afterSeq > 0 {
		out := make([]TranscriptRow, 0, len(rows))
		for _, row := range rows {
			if row.Seq > afterSeq {
				out = append(out, row)
			}
		}
		return out
	}
	if len(rows) > tail {
		return rows[len(rows)-tail:]
	}
	return rows
}

func transcriptRow(event Event) (TranscriptRow, bool) {
	switch event.Type {
	case "user/message":
		if userSourceKind(event.Data) != "user" {
			return TranscriptRow{}, false
		}
		text, err := messageText(event)
		if err != nil || strings.TrimSpace(text) == "" {
			return TranscriptRow{}, false
		}
		return TranscriptRow{Seq: event.Seq, Time: event.Time, Kind: TranscriptUser, Text: text}, true
	case "assistant/message":
		text, err := messageText(event)
		if err != nil || strings.TrimSpace(text) == "" {
			return TranscriptRow{}, false
		}
		return TranscriptRow{Seq: event.Seq, Time: event.Time, Kind: TranscriptAssistant, Text: text}, true
	case "tool/call", "command/run":
		name, text := toolCallPreview(event)
		if name == "" && text == "" {
			return TranscriptRow{}, false
		}
		return TranscriptRow{Seq: event.Seq, Time: event.Time, Kind: TranscriptTool, Text: text, Tool: name}, true
	default:
		return TranscriptRow{}, false
	}
}

func userSourceKind(raw json.RawMessage) string {
	var data struct {
		Source struct {
			Kind string `json:"kind"`
		} `json:"source"`
	}
	if json.Unmarshal(raw, &data) != nil {
		return ""
	}
	return strings.TrimSpace(data.Source.Kind)
}

func messageText(event Event) (string, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return "", err
	}
	if data == nil {
		return "", nil
	}
	if event.Type == "assistant/message" {
		raw := data["message"]
		data = nil
		if err := json.Unmarshal(raw, &data); err != nil || data == nil {
			return "", err
		}
	}
	var blocks []map[string]json.RawMessage
	if err := json.Unmarshal(data["content"], &blocks); err != nil || blocks == nil {
		return "", err
	}
	var text strings.Builder
	for _, block := range blocks {
		var kind string
		if json.Unmarshal(block["type"], &kind) != nil || kind != "text" {
			continue
		}
		var value *string
		if json.Unmarshal(block["text"], &value) != nil || value == nil {
			continue
		}
		text.WriteString(*value)
	}
	return text.String(), nil
}

func toolCallPreview(event Event) (string, string) {
	name := toolName(event)
	var data map[string]json.RawMessage
	if json.Unmarshal(event.Data, &data) != nil || data == nil {
		return name, ""
	}
	var args map[string]json.RawMessage
	if json.Unmarshal(data["arguments"], &args) != nil || args == nil {
		return name, ""
	}
	var command string
	if json.Unmarshal(args["command"], &command) == nil {
		return name, strings.TrimSpace(command)
	}
	return name, ""
}
