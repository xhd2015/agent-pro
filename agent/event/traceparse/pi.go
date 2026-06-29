package traceparse

import (
	"encoding/json"
	"fmt"
	"strings"

	pi_types "github.com/xhd2015/agent-pro/agent/event/pi_types"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type piTraceAdapter struct{}

func init() {
	RegisterAdapter(12, piTraceAdapter{})
}

func (piTraceAdapter) Name() string {
	return "pi"
}

func (piTraceAdapter) Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	var event pi_types.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	switch event.Type {
	case pi_types.EventTypeSession:
		return traceview.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeAgentStart, pi_types.EventTypeAgentEnd:
		return traceview.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeTurnStart, pi_types.EventTypeTurnEnd:
		return traceview.AgentTraceParsedEvent{}, false
	case pi_types.EventTypeMessageStart:
		if event.Message != nil && event.Message.Role == "assistant" {
			text := piExtractText(event.Message)
			if text != "" {
				return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
					Role:    traceview.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeMessageUpdate:
		if event.Message != nil && event.Message.Role == "assistant" {
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
				return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
					Role:    traceview.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeMessageEnd:
		if event.Message != nil && event.Message.Role == "assistant" {
			text := ""
			if event.AssistantMessageEvent != nil {
				text = event.AssistantMessageEvent.Delta
				if strings.TrimSpace(text) == "" {
					text = ""
				}
			}
			if text != "" {
				return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
					Role:    traceview.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case pi_types.EventTypeToolExecStart:
		activity := piToolExecStart(event)
		if activity != nil {
			return traceview.AgentTraceParsedEvent{Activity: activity}, true
		}
	case pi_types.EventTypeToolExecEnd:
		activity := piToolExecEnd(event)
		if activity != nil {
			return traceview.AgentTraceParsedEvent{Activity: activity}, true
		}
	}
	return traceview.AgentTraceParsedEvent{}, false
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

func piToolExecStart(event pi_types.Event) *traceview.AgentTraceActivity {
	summaryText := ""
	if event.Args != nil {
		if cmd, ok := event.Args["command"].(string); ok {
			summaryText = cmd
		}
		if summaryText == "" {
			data, err := json.Marshal(event.Args)
			if err == nil {
				summaryText = string(data)
			}
		}
	}
	toolName := event.ToolName
	if toolName == "" {
		toolName = "Tool"
	}
	return &traceview.AgentTraceActivity{
		Subtype:  traceview.SubtypeStarted,
		ToolName: toolName,
		Summary:  summaryText,
		Status:   traceview.StatusInProgress,
	}
}

func piToolExecEnd(event pi_types.Event) *traceview.AgentTraceActivity {
	summaryText := piResultToString(event.Result)
	toolName := event.ToolName
	if toolName == "" {
		toolName = "Tool"
	}
	status := traceview.StatusCompleted
	if event.IsError {
		status = traceview.StatusFailed
	}
	return &traceview.AgentTraceActivity{
		Subtype:  traceview.SubtypeCompleted,
		ToolName: toolName,
		Summary:  summaryText,
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