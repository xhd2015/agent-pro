package agent_trace

import (
	"encoding/json"
	"strings"
)

type cursorTraceAdapter struct{}

func init() {
	RegisterAgentTraceAdapter(20, cursorTraceAdapter{})
}

func (cursorTraceAdapter) Name() string {
	return "cursor"
}

func (cursorTraceAdapter) Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	var event traceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return AgentTraceParsedEvent{}, false
	}
	if event.Type != "tool_call" {
		return AgentTraceParsedEvent{}, false
	}
	if activity := traceCursorActivity(event); activity != nil {
		return AgentTraceParsedEvent{Activity: activity}, true
	}
	return AgentTraceParsedEvent{}, false
}

func traceCursorActivity(event traceEvent) *AgentTraceActivity {
	toolName, summary := traceCursorToolInfo(event.ToolCall, event.Subtype)
	if toolName == "" {
		toolName = "Tool"
	}
	status := ""
	switch event.Subtype {
	case "completed":
		status = "completed"
	case "started", "updated":
		status = "in_progress"
	}
	return &AgentTraceActivity{
		Subtype:  event.Subtype,
		CallID:   event.CallID,
		ToolName: toolName,
		Summary:  summary,
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
			name = titleFromIdentifier(key)
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
		return compactTraceOutput(string(data))
	}
	for _, key := range []string{"command", "path", "pattern", "query", "description"} {
		if value, ok := parsed.Args[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
