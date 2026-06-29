package traceparse

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-pro/agent/event/summary"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type opencodeTraceAdapter struct{}

func init() {
	RegisterAdapter(15, opencodeTraceAdapter{})
}

func (opencodeTraceAdapter) Name() string {
	return "opencode"
}

type OpencodeRunLine struct {
	Type      string           `json:"type"`
	Timestamp int64            `json:"timestamp,omitempty"`
	SessionID string           `json:"sessionID,omitempty"`
	Part      *OpencodeRunPart `json:"part,omitempty"`
	Error     map[string]any   `json:"error,omitempty"`
}

type OpencodeRunPart struct {
	ID     string             `json:"id"`
	Type   string             `json:"type"`
	Tool   string             `json:"tool,omitempty"`
	Text   string             `json:"text,omitempty"`
	CallID string             `json:"callID,omitempty"`
	State  *OpencodePartState `json:"state,omitempty"`
}

type OpencodePartState struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Title  string         `json:"title,omitempty"`
	Output string         `json:"output,omitempty"`
	Input  map[string]any `json:"input,omitempty"`
}

func (opencodeTraceAdapter) Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	var line OpencodeRunLine
	if err := json.Unmarshal(raw, &line); err != nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if line.Type == "" {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if line.Part == nil && line.Type != "error" {
		return traceview.AgentTraceParsedEvent{}, false
	}

	switch line.Type {
	case "text", "reasoning":
		if line.Part != nil {
			text := strings.TrimSpace(line.Part.Text)
			if text != "" {
				return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
					Role:    traceview.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case "tool_use":
		if activity := OpencodeToolActivity(line.Part); activity != nil {
			return traceview.AgentTraceParsedEvent{Activity: activity}, true
		}
	case "error":
		return opencodeErrorEvent(line.Error), true
	}
	return traceview.AgentTraceParsedEvent{}, false
}

func OpencodeToolActivity(part *OpencodeRunPart) *traceview.AgentTraceActivity {
	if part == nil || part.Type != "tool" {
		return nil
	}
	subtype := traceview.SubtypeStarted
	status := traceview.StatusInProgress
	summaryText := ""

	if part.State != nil {
		switch part.State.Status {
		case "pending":
			subtype = traceview.SubtypeStarted
			status = traceview.StatusPending
		case "running":
			subtype = traceview.SubtypeStarted
			status = traceview.StatusInProgress
		case "completed":
			subtype = traceview.SubtypeCompleted
			status = traceview.StatusCompleted
		case "error":
			subtype = traceview.SubtypeCompleted
			status = traceview.StatusFailed
		}

		var parts []string

		if part.State.Input != nil {
			if inputSummary := summary.ToolInputSummary(part.Tool, part.State.Input); inputSummary != "" {
				parts = append(parts, inputSummary)
			}
		}

		if part.State.Title != "" {
			parts = append(parts, part.State.Title)
		}

		if part.State.Output != "" {
			parts = append(parts, summary.CompactTraceOutput(part.State.Output))
		}

		if part.State.Error != "" {
			summaryText = summary.CompactTraceOutput(part.State.Error)
		} else {
			summaryText = strings.Join(parts, "\n")
			if summaryText == "" && part.State.Input != nil {
				if path, ok := part.State.Input["path"].(string); ok {
					summaryText = path
				} else if pattern, ok := part.State.Input["pattern"].(string); ok {
					summaryText = pattern
				} else if query, ok := part.State.Input["query"].(string); ok {
					summaryText = query
				}
			}
		}
	}

	toolName := part.Tool
	if toolName == "" {
		toolName = "Tool"
	}
	friendly := opencodeFriendlyName(toolName)

	return &traceview.AgentTraceActivity{
		Subtype:  subtype,
		CallID:   part.CallID,
		ToolName: friendly,
		Summary:  summaryText,
		Kind:     part.Tool,
		Status:   status,
	}
}

func opencodeFriendlyName(tool string) string {
	switch tool {
	case "bash":
		return "Shell"
	case "Read", "read":
		return "Read File"
	case "Edit", "edit":
		return "Edit File"
	case "Write", "write":
		return "Write File"
	case "Glob", "glob":
		return "Glob"
	case "Grep", "grep":
		return "Grep"
	case "WebSearch", "websearch":
		return "Web Search"
	case "WebFetch", "webfetch":
		return "Web Fetch"
	case "Task", "task":
		return "Sub-Agent"
	case "TodoWrite", "todowrite":
		return "Plan"
	case "LSP", "lsp":
		return "LSP"
	case "Skill", "skill":
		return "Skill"
	default:
		return tool
	}
}

func opencodeErrorEvent(errData map[string]any) traceview.AgentTraceParsedEvent {
	errMsg := ""
	if errData != nil {
		if name, ok := errData["name"].(string); ok {
			errMsg = name
		}
		if data, ok := errData["data"].(map[string]any); ok {
			if msg, ok := data["message"].(string); ok && msg != "" {
				errMsg = msg
			}
		}
	}
	if errMsg == "" {
		errMsg = "opencode error"
	}
	return traceview.AgentTraceParsedEvent{Activity: &traceview.AgentTraceActivity{
		Subtype:  traceview.SubtypeCompleted,
		ToolName: "Codex",
		Summary:  errMsg,
		Kind:     "runtime_error",
		Status:   traceview.StatusFailed,
	}}
}