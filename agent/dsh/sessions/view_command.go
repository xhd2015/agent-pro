package sessions

import (
	"encoding/json"
	"strings"
)

const TranscriptCommand = "command"

func commandID(event Event) string {
	var data struct {
		CommandID string `json:"commandId"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return ""
	}
	return strings.TrimSpace(data.CommandID)
}

func commandRunMessage(event Event) Message {
	name := toolName(event)
	return Message{
		Seq:  event.Seq,
		Time: event.Time,
		Kind: TranscriptCommand,
		Tool: name,
	}
}

func commandDoneMessage(event Event) Message {
	msg := commandRunMessage(event)
	applyCommandDone(&msg, event)
	return msg
}

func applyCommandDone(row *Message, event Event) {
	var data struct {
		Kind string `json:"kind"`
		Text string `json:"text"`
	}
	if json.Unmarshal(event.Data, &data) != nil {
		return
	}
	row.Text = strings.TrimSpace(data.Text)
	row.Sum = strings.TrimSpace(data.Kind)
}
