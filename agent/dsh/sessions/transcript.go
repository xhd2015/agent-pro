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
	// Deleted marks a row an edit removed while folding. It never reaches a
	// reader: the fold drops these before returning.
	Deleted bool `json:"-"`
}

// Transcript projects committed user, assistant, and tool/call events.
// afterSeq > 0 returns only later seqs (no tail trim). afterSeq == 0 keeps
// the last tail rows (tail <= 0 means 24).
func Transcript(events []Event, afterSeq int64, tail int) []TranscriptRow {
	return TranscriptFilter(TranscriptRows(events), afterSeq, tail)
}

// ApplyEdit folds one message/edit into rows a streaming reader already holds.
//
// A streaming reader appends as it goes, so it cannot see the message an edit
// names unless it keeps every row; this is the entry that lets it apply the edit
// to what it has. A target it never saw is skipped.
func ApplyEdit(rows []TranscriptRow, event Event) []TranscriptRow {
	if event.Type != "message/edit" {
		return rows
	}
	return applyMessageEdits(rows, event)
}

// HasEdit reports whether the events contain a message/edit.
func HasEdit(events []Event) bool {
	for _, event := range events {
		if event.Type == "message/edit" {
			return true
		}
	}
	return false
}

// TranscriptRows folds events into transcript rows without trimming.
//
// Two passes, because an edit arrives after the message it changes: the first
// pass produces the base rows, the second applies every committed edit. A row
// keeps the seq of the message it came from, so a reader that keys by seq still
// matches the rows it already holds after an edit.
func TranscriptRows(events []Event) []TranscriptRow {
	out := make([]TranscriptRow, 0, len(events))
	for _, event := range events {
		row, ok := transcriptRow(event)
		if !ok {
			continue
		}
		out = append(out, row)
	}
	for _, event := range events {
		if event.Type != "message/edit" {
			continue
		}
		out = applyMessageEdits(out, event)
	}
	return out
}

// applyMessageEdits rewrites or drops the rows one message/edit names.
//
// A target that is not in the current rows is skipped rather than reported: a
// later compaction can shadow the node an older edit names, and refusing would
// turn a committed log into an unreadable one.
func applyMessageEdits(rows []TranscriptRow, event Event) []TranscriptRow {
	var data struct {
		Edits []struct {
			Seq     int64  `json:"seq"`
			Text    string `json:"text"`
			Deleted bool   `json:"deleted"`
		} `json:"edits"`
	}
	if json.Unmarshal(event.Data, &data) != nil || len(data.Edits) == 0 {
		return rows
	}
	byIndex := make(map[int64]int, len(rows))
	for i, row := range rows {
		byIndex[row.Seq] = i
	}
	for _, edit := range data.Edits {
		index, ok := byIndex[edit.Seq]
		if !ok {
			continue
		}
		if edit.Deleted {
			// Mark for removal; the caller compacts below so indexes stay valid.
			rows[index].Deleted = true
			continue
		}
		rows[index].Text = edit.Text
	}
	kept := rows[:0]
	for _, row := range rows {
		if !row.Deleted {
			kept = append(kept, row)
		}
	}
	return kept
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
	// toolArguments decodes both shapes a transcript can hold: a plain JSON
	// object, and the JSON string some runners double-encode it into. Parsing
	// the arguments straight into a map dropped every tool/call row whose
	// arguments arrived as a string, which is most of them.
	args := toolArguments(event)
	if args == nil {
		return name, ""
	}
	var command string
	if json.Unmarshal(args["command"], &command) == nil {
		return name, strings.TrimSpace(command)
	}
	return name, ""
}
