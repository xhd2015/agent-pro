package fakeagent

import (
	"encoding/json"
	"fmt"
	"strings"
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

func FormatCodexEventsText(events []Event) string {
	var b strings.Builder
	for _, e := range events {
		if e.Item == nil || e.Type != EventCompleted {
			continue
		}
		switch e.Item.Type {
		case ItemReasoning:
			if text := strings.TrimSpace(e.Item.Text); text != "" {
				b.WriteString("\nThinking...\n")
				for _, line := range strings.Split(text, "\n") {
					b.WriteString("  " + line + "\n")
				}
			}
		case ItemCommandExecution:
			cmd := e.Item.Command
			output := strings.TrimSpace(e.Item.AggregatedOutput)
			b.WriteString("\n> " + cmd + "\n")
			if output != "" {
				for _, line := range strings.Split(output, "\n") {
					b.WriteString("  " + line + "\n")
				}
			}
		case ItemFileChange:
			for _, c := range e.Item.Changes {
				symbol := "+"
				verb := "created"
				switch c.Kind {
				case "modify":
					symbol = "~"
					verb = "modified"
				case "delete":
					symbol = "-"
					verb = "deleted"
				}
				b.WriteString(symbol + " " + c.Path + " (" + verb + ")\n")
			}
		case ItemMessage:
			if text := strings.TrimSpace(e.Item.Text); text != "" {
				b.WriteString("\n" + text + "\n")
			}
		}
	}
	return b.String()
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
