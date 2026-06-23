package opencode

import (
	"encoding/json"
	"strings"

	"github.com/xhd2015/agent-pro/agent_trace/types"
)

type opencodeTraceAdapter struct{}

func init() {
	types.RegisterAgentTraceAdapter(15, opencodeTraceAdapter{})
}

func (opencodeTraceAdapter) Name() string {
	return "opencode"
}

type opencodeRunLine struct {
	Type      string           `json:"type"`
	Timestamp int64            `json:"timestamp,omitempty"`
	SessionID string           `json:"sessionID,omitempty"`
	Part      *opencodeRunPart `json:"part,omitempty"`
	Error     map[string]any   `json:"error,omitempty"`
}

type opencodeRunPart struct {
	ID     string             `json:"id"`
	Type   string             `json:"type"`
	Tool   string             `json:"tool,omitempty"`
	Text   string             `json:"text,omitempty"`
	CallID string             `json:"callID,omitempty"`
	State  *opencodePartState `json:"state,omitempty"`
}

type opencodePartState struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Title  string         `json:"title,omitempty"`
	Output string         `json:"output,omitempty"`
	Input  map[string]any `json:"input,omitempty"`
}

func (opencodeTraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
	var line opencodeRunLine
	if err := json.Unmarshal(raw, &line); err != nil {
		return types.AgentTraceParsedEvent{}, false
	}
	if line.Type == "" {
		return types.AgentTraceParsedEvent{}, false
	}
	if line.Part == nil && line.Type != "error" {
		return types.AgentTraceParsedEvent{}, false
	}

	switch line.Type {
	case "text", "reasoning":
		if line.Part != nil {
			text := strings.TrimSpace(line.Part.Text)
			if text != "" {
				return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
					Role:    types.RoleAssistant,
					Content: text,
				}}, true
			}
		}
	case "tool_use":
		if activity := opencodeToolActivity(line.Part); activity != nil {
			return types.AgentTraceParsedEvent{Activity: activity}, true
		}
	case "error":
		return opencodeErrorEvent(line.Error), true
	}
	return types.AgentTraceParsedEvent{}, false
}

func opencodeToolActivity(part *opencodeRunPart) *types.AgentTraceActivity {
	if part == nil || part.Type != "tool" {
		return nil
	}
	subtype := types.SubtypeStarted
	status := types.StatusInProgress
	summary := ""

	if part.State != nil {
		switch part.State.Status {
		case "pending":
			subtype = types.SubtypeStarted
			status = types.StatusPending
		case "running":
			subtype = types.SubtypeStarted
			status = types.StatusInProgress
		case "completed":
			subtype = types.SubtypeCompleted
			status = types.StatusCompleted
		case "error":
			subtype = types.SubtypeCompleted
			status = types.StatusFailed
		}

		var parts []string

		if part.State.Input != nil {
			if inputSummary := opencodeToolInputSummary(part.Tool, part.State.Input); inputSummary != "" {
				parts = append(parts, inputSummary)
			}
		}

		if part.State.Title != "" {
			parts = append(parts, part.State.Title)
		}

		if part.State.Output != "" {
			parts = append(parts, types.CompactTraceOutput(part.State.Output))
		}

		if part.State.Error != "" {
			summary = types.CompactTraceOutput(part.State.Error)
		} else {
			summary = strings.Join(parts, "\n")
			if summary == "" && part.State.Input != nil {
				if path, ok := part.State.Input["path"].(string); ok {
					summary = path
				} else if pattern, ok := part.State.Input["pattern"].(string); ok {
					summary = pattern
				} else if query, ok := part.State.Input["query"].(string); ok {
					summary = query
				}
			}
		}
	}

	toolName := part.Tool
	if toolName == "" {
		toolName = "Tool"
	}
	friendly := opencodeFriendlyName(toolName)

	return &types.AgentTraceActivity{
		Subtype:  subtype,
		CallID:   part.CallID,
		ToolName: friendly,
		Summary:  summary,
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

func opencodeToolInputSummary(tool string, input map[string]any) string {
	if len(input) == 0 {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(tool)) {
	case "skill":
		return opencodeSkillInputSummary(input)
	case "todowrite":
		return opencodeTodoWriteInputSummary(input)
	default:
		return stringInputValue(input, "command")
	}
}

func opencodeSkillInputSummary(input map[string]any) string {
	var parts []string
	if name := stringInputValue(input, "skill", "name"); name != "" {
		parts = append(parts, name)
	}
	if args := stringInputValue(input, "arguments", "args", "command"); args != "" {
		parts = append(parts, args)
	}
	return strings.Join(parts, "\n")
}

func opencodeTodoWriteInputSummary(input map[string]any) string {
	rawTodos, ok := input["todos"]
	if !ok {
		return ""
	}
	todos, ok := rawTodos.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, rawTodo := range todos {
		todo, ok := rawTodo.(map[string]any)
		if !ok {
			continue
		}
		content := stringInputValue(todo, "content", "text", "task")
		if content == "" {
			continue
		}
		if status := stringInputValue(todo, "status"); status != "" {
			parts = append(parts, status+": "+content)
		} else {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

func stringInputValue(input map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := input[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		if text = strings.TrimSpace(text); text != "" {
			return text
		}
	}
	return ""
}

func opencodeErrorEvent(errData map[string]any) types.AgentTraceParsedEvent {
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
	return types.AgentTraceParsedEvent{Activity: &types.AgentTraceActivity{
		Subtype:  types.SubtypeCompleted,
		ToolName: "Codex",
		Summary:  errMsg,
		Kind:     "runtime_error",
		Status:   types.StatusFailed,
	}}
}
