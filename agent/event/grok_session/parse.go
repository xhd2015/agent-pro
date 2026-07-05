package grok_session

import (
	"encoding/json"
	"strings"
)

type textBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// TextContent extracts user-visible text from a session update content field.
func TextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var block textBlock
	if err := json.Unmarshal(raw, &block); err == nil && block.Text != "" {
		return block.Text
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	return ""
}

type wireEnvelope struct {
	Method string `json:"method"`
	Params struct {
		SessionID string          `json:"sessionId"`
		Update    json.RawMessage `json:"update"`
	} `json:"params"`
}

// ParseLine parses one updates.jsonl line in flat or grok wire envelope form.
func ParseLine(line string) (SessionUpdate, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return SessionUpdate{}, false
	}

	var flat SessionUpdate
	if err := json.Unmarshal([]byte(line), &flat); err == nil && strings.TrimSpace(flat.SessionUpdate) != "" {
		return flat, true
	}

	var wire wireEnvelope
	if err := json.Unmarshal([]byte(line), &wire); err != nil {
		return SessionUpdate{}, false
	}
	if len(wire.Params.Update) == 0 {
		return SessionUpdate{}, false
	}
	var nested SessionUpdate
	if err := json.Unmarshal(wire.Params.Update, &nested); err != nil {
		return SessionUpdate{}, false
	}
	if strings.TrimSpace(nested.SessionUpdate) == "" {
		return SessionUpdate{}, false
	}
	return nested, true
}