package fakeagent

import (
	"encoding/json"
	"fmt"
)

func FormatCodexEvents(events []Event) ([]string, error) {
	lines := make([]string, 0, len(events))
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return nil, fmt.Errorf("marshal codex event: %w", err)
		}
		lines = append(lines, string(data))
	}
	return lines, nil
}

type CodexEvent struct {
	Type    string     `json:"type"`
	Item    *EventItem `json:"item,omitempty"`
	Delta   string     `json:"delta,omitempty"`
	Text    string     `json:"text,omitempty"`
	Message string     `json:"message,omitempty"`
}

func ParseCodexEvents(lines []string) ([]Event, error) {
	events := make([]Event, 0, len(lines))
	for i, line := range lines {
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("line %d: parse codex event: %w", i+1, err)
		}
		events = append(events, event)
	}
	return events, nil
}
