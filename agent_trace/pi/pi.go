package pi

import (
	"encoding/json"
	"fmt"
	"strings"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	"github.com/xhd2015/agent-pro/agent_trace/types"
)

type piTraceAdapter struct{}

func init() {
	types.RegisterAgentTraceAdapter(12, piTraceAdapter{})
}

func (piTraceAdapter) Name() string {
	return "pi"
}

func (piTraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
	var event pi_types.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		return types.AgentTraceParsedEvent{}, false
	}
	switch event.Type {
	case pi_types.EventTypeSession:
		return types.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeAgentStart, pi_types.EventTypeAgentEnd:
		return types.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeTurnStart, pi_types.EventTypeTurnEnd:
		return types.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeMessageStart:
		if event.Message != nil && event.Message.Role == "assistant" {
			text := piExtractText(event.Message)
			if text != "" {
				return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
					Role:    types.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeMessageUpdate:
		if event.Message != nil && event.Message.Role == "assistant" {
			// Prefer delta over accumulated Content text (streaming UX)
			text := ""
			if event.AssistantMessageEvent != nil {
				text = event.AssistantMessageEvent.Delta
				if strings.TrimSpace(text) == "" {
					text = ""
				}
			}
			if text == "" {
				text = piExtractText(event.Message)
			}
			if text != "" {
				return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
					Role:    types.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeMessageEnd:
		if event.Message != nil && event.Message.Role == "assistant" {
			// Prefer delta over full Content text; deltas are already shown via message_update.
			// Don't output full accumulated text again to prevent duplication.
			text := ""
			if event.AssistantMessageEvent != nil {
				text = event.AssistantMessageEvent.Delta
				if strings.TrimSpace(text) == "" {
					text = ""
				}
			}
			if text != "" {
				return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
					Role:    types.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeToolExecStart:
		activity := piToolExecStart(event)
		if activity != nil {
			return types.AgentTraceParsedEvent{Activity: activity}, true
		}
	case pi_types.EventTypeToolExecEnd:
		activity := piToolExecEnd(event)
		if activity != nil {
			return types.AgentTraceParsedEvent{Activity: activity}, true
		}
	}
	return types.AgentTraceParsedEvent{}, false
}

func piExtractText(msg *pi_types.AgentMessage) string {
	if msg == nil {
		return ""
	}
	for _, c := range msg.Content {
		if c.Text != "" {
			return c.Text
		}
		if c.Thinking != "" {
			return c.Thinking
		}
	}
	return ""
}

func piToolExecStart(event pi_types.Event) *types.AgentTraceActivity {
	summary := ""
	if event.Args != nil {
		if cmd, ok := event.Args["command"].(string); ok {
			summary = cmd
		}
		if summary == "" {
			data, err := json.Marshal(event.Args)
			if err == nil {
				summary = string(data)
			}
		}
	}
	toolName := event.ToolName
	if toolName == "" {
		toolName = "Tool"
	}
	return &types.AgentTraceActivity{
		Subtype:  types.SubtypeStarted,
		ToolName: toolName,
		Summary:  summary,
		Status:   types.StatusInProgress,
	}
}

func piToolExecEnd(event pi_types.Event) *types.AgentTraceActivity {
	summary := piResultToString(event.Result)
	toolName := event.ToolName
	if toolName == "" {
		toolName = "Tool"
	}
	status := types.StatusCompleted
	if event.IsError {
		status = types.StatusFailed
	}
	return &types.AgentTraceActivity{
		Subtype:  types.SubtypeCompleted,
		ToolName: toolName,
		Summary:  summary,
		Status:   status,
	}
}

func piResultToString(result any) string {
	if result == nil {
		return ""
	}
	if s, ok := result.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", result)
}
