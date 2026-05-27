package cursor

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-traces/agent_trace/types"
)

type cursorTraceAdapter struct{}

func init() {
	types.RegisterAgentTraceAdapter(20, cursorTraceAdapter{})
}

func (cursorTraceAdapter) Name() string {
	return "cursor"
}

func (cursorTraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
	var event types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return types.AgentTraceParsedEvent{}, false
	}
	if event.Type != "tool_call" {
		return types.AgentTraceParsedEvent{}, false
	}
	if activity := traceCursorActivity(event); activity != nil {
		return types.AgentTraceParsedEvent{Activity: activity}, true
	}
	return types.AgentTraceParsedEvent{}, false
}

func traceCursorActivity(event types.TraceEvent) *types.AgentTraceActivity {
	toolName, summary := traceCursorToolInfo(event.ToolCall, event.Subtype)
	if toolName == "" {
		toolName = "Tool"
	}
	var status types.ActivityStatus
	switch event.Subtype {
	case "completed":
		status = types.StatusCompleted
	case "started", "updated":
		status = types.StatusInProgress
	}
	return &types.AgentTraceActivity{
		Subtype:  types.ActivitySubtype(event.Subtype),
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
			name = types.TitleFromIdentifier(key)
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
		return types.CompactTraceOutput(string(data))
	}
	for _, key := range []string{"command", "path", "pattern", "query", "description"} {
		if value, ok := parsed.Args[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
