package agentui

import (
	"fmt"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

// BuildContinuationPrompt formats prior timeline messages plus the new user turn for the runner.
func BuildContinuationPrompt(prior []types.AgentEvent, newPrompt string) string {
	newPrompt = strings.TrimSpace(newPrompt)
	events := prior
	if len(events) > 0 {
		last := events[len(events)-1]
		if last.Type == types.ActionMessage &&
			strings.TrimSpace(last.Role) == "user" &&
			strings.TrimSpace(last.Text) == newPrompt {
			events = events[:len(events)-1]
		}
	}

	var lines []string
	for _, ev := range events {
		if ev.Type != types.ActionMessage {
			continue
		}
		role := strings.TrimSpace(ev.Role)
		text := strings.TrimSpace(ev.Text)
		if text == "" {
			continue
		}
		switch role {
		case "user":
			lines = append(lines, fmt.Sprintf("User: %s", text))
		case "assistant":
			lines = append(lines, fmt.Sprintf("Assistant: %s", text))
		default:
			continue
		}
	}

	if len(lines) == 0 {
		return newPrompt
	}

	var b strings.Builder
	b.WriteString("Previous conversation:\n")
	b.WriteString(strings.Join(lines, "\n"))
	b.WriteString("\n\nUser: ")
	b.WriteString(newPrompt)
	return b.String()
}