package traceparse

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-pro/agent/event/cursor_types"
	"github.com/xhd2015/agent-pro/agent/event/summary"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type cursorTraceAdapter struct{}

func init() {
	RegisterAdapter(20, cursorTraceAdapter{})
}

func (cursorTraceAdapter) Name() string {
	return "cursor"
}

func (cursorTraceAdapter) Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	var event cursor_types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if event.Type != "tool_call" {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if activity := traceCursorActivity(event); activity != nil {
		return traceview.AgentTraceParsedEvent{Activity: activity}, true
	}
	return traceview.AgentTraceParsedEvent{}, false
}

func traceCursorActivity(event cursor_types.TraceEvent) *traceview.AgentTraceActivity {
	toolName, summaryText := traceCursorToolInfo(event.ToolCall, event.Subtype)
	if toolName == "" {
		toolName = "Tool"
	}
	var status traceview.ActivityStatus
	switch event.Subtype {
	case "completed":
		status = traceview.StatusCompleted
	case "started", "updated":
		status = traceview.StatusInProgress
	}
	return &traceview.AgentTraceActivity{
		Subtype:  traceview.ActivitySubtype(event.Subtype),
		CallID:   event.CallID,
		ToolName: toolName,
		Summary:  summaryText,
		Status:   status,
	}
}

var traceCursorToolNames = map[string]string{
	"readToolCall":        "Read File",
	"shellToolCall":       "Shell",
	"writeToolCall":       "Write File",
	"editToolCall":        "Edit File",
	"searchToolCall":      "Search",
	"grepToolCall":        "Grep",
	"globToolCall":        "Glob",
	"updateTodosToolCall": "Update Todos",
	"mcpToolCall":         "MCP Tool",
	"taskToolCall":        "Sub-Agent",
	"listToolCall":        "List Files",
}

func traceCursorToolInfo(toolCall map[string]json.RawMessage, subtype string) (string, string) {
	for key, raw := range toolCall {
		name := traceCursorToolNames[key]
		if name == "" {
			name = summary.TitleFromIdentifier(key)
		}
		return name, traceCursorToolSummary(raw, subtype)
	}
	return "", ""
}

func traceCursorToolSummary(raw json.RawMessage, subtype string) string {
	var parsed struct {
		Args   map[string]any `json:"args"`
		Result any            `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ""
	}
	if subtype == "completed" && parsed.Result != nil {
		data, err := json.Marshal(parsed.Result)
		if err != nil {
			return "[completed]"
		}
		return summary.CompactTraceOutput(string(data))
	}
	for _, key := range []string{"command", "path", "pattern", "query", "description"} {
		if value, ok := parsed.Args[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}