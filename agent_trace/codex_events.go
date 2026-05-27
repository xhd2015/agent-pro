package agent_trace

import (
	"encoding/json"
	"fmt"
	"strings"
)

type codexTraceAdapter struct{}

func init() {
	RegisterAgentTraceAdapter(10, codexTraceAdapter{})
}

func (codexTraceAdapter) Name() string {
	return "codex"
}

func (codexTraceAdapter) Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	var event traceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return AgentTraceParsedEvent{}, false
	}
	if event.Item == nil {
		return AgentTraceParsedEvent{}, false
	}
	if text := traceCodexAssistantText(event); text != "" {
		return AgentTraceParsedEvent{Message: &AgentTraceMessage{
			Role:    "assistant",
			Content: text,
		}}, true
	}
	if activity := traceCodexActivity(event); activity != nil {
		return AgentTraceParsedEvent{Activity: activity}, true
	}
	return AgentTraceParsedEvent{}, false
}

func traceCodexAssistantText(event traceEvent) string {
	if event.Type != "item.completed" || event.Item == nil || !traceIsAssistantItem(event.Item.Type) {
		return ""
	}
	return traceItemText(event.Item)
}

const traceCodexHooksDeprecatedPrefix = "`[features].codex_hooks` is deprecated."

func traceCodexActivity(event traceEvent) *AgentTraceActivity {
	if event.Item == nil || traceIsAssistantItem(event.Item.Type) {
		return nil
	}
	subtype := ""
	switch event.Type {
	case "item.started":
		subtype = "started"
	case "item.updated":
		subtype = "updated"
	case "item.completed":
		subtype = "completed"
	default:
		return nil
	}
	summary, replace := traceCodexSummary(event.Item, subtype)
	kind := event.Item.Type
	toolName := traceCodexToolName(kind)
	if isTraceCodexHooksDeprecation(summary) {
		kind = "warning"
		toolName = "Config Warning"
	}
	return &AgentTraceActivity{
		Subtype:        subtype,
		CallID:         event.Item.ID,
		ToolName:       toolName,
		Summary:        summary,
		Kind:           kind,
		Status:         traceCodexStatus(event.Item, subtype, summary),
		FileChanges:    event.Item.Changes,
		ReplaceSummary: replace,
	}
}

func traceCodexStatus(item *traceItem, subtype, summary string) string {
	if item == nil {
		return ""
	}
	if status := strings.TrimSpace(item.Status); status != "" {
		return status
	}
	if item.Type == "error" {
		if isTraceCodexHooksDeprecation(summary) {
			return "warning"
		}
		return "failed"
	}
	if subtype == "completed" {
		if item.ExitCode != nil && *item.ExitCode != 0 {
			return "failed"
		}
		return "completed"
	}
	return "in_progress"
}

func isTraceCodexHooksDeprecation(summary string) bool {
	return strings.HasPrefix(strings.TrimSpace(summary), traceCodexHooksDeprecatedPrefix)
}

func traceCodexSummary(item *traceItem, subtype string) (string, bool) {
	switch item.Type {
	case "command_execution":
		if subtype == "started" {
			return item.Command, false
		}
		return summarizeTraceCommandOutput(item), true
	case "reasoning":
		return traceItemText(item), subtype != "started"
	case "todo_list", "todo_write":
		return summarizeTraceTodoList(item.Items), subtype != "started"
	case "plan_update":
		return summarizeTracePlanUpdate(item), subtype != "started"
	case "file_change":
		return summarizeTraceFileChanges(item.Changes), subtype != "started"
	default:
		text := traceItemText(item)
		if text != "" {
			return text, subtype != "started"
		}
		if item.Command != "" {
			return item.Command, subtype != "started"
		}
		if item.Status != "" {
			return item.Status, subtype != "started"
		}
		return "", subtype != "started"
	}
}

func summarizeTraceCommandOutput(item *traceItem) string {
	var b strings.Builder
	if item.Command != "" {
		b.WriteString(item.Command)
	}
	if item.ExitCode != nil {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("[exit %d]", *item.ExitCode))
	}
	output := strings.TrimSpace(item.AggregatedOutput)
	if output != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(compactTraceOutput(output))
	}
	return b.String()
}

func summarizeTraceTodoList(items []traceTodoItem) string {
	if len(items) == 0 {
		return ""
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}
		marker := "[ ]"
		if item.Completed || strings.EqualFold(item.Status, "completed") || strings.EqualFold(item.Status, "done") {
			marker = "[x]"
		} else if item.Status != "" && !strings.EqualFold(item.Status, "pending") {
			marker = "[" + strings.TrimSpace(item.Status) + "]"
		}
		lines = append(lines, marker+" "+text)
	}
	return strings.Join(lines, "\n")
}

func summarizeTracePlanUpdate(item *traceItem) string {
	var lines []string
	if explanation := strings.TrimSpace(item.Explanation); explanation != "" {
		lines = append(lines, explanation)
	}
	for _, step := range item.Plan {
		text := strings.TrimSpace(step.Step)
		if text == "" {
			continue
		}
		status := strings.TrimSpace(step.Status)
		if status == "" {
			lines = append(lines, text)
			continue
		}
		lines = append(lines, "["+status+"] "+text)
	}
	return strings.Join(lines, "\n")
}

func summarizeTraceFileChanges(changes []FileChange) string {
	if len(changes) == 0 {
		return ""
	}
	lines := make([]string, 0, len(changes)+1)
	lines = append(lines, fmt.Sprintf("%d file%s changed", len(changes), pluralSuffix(len(changes))))
	for _, change := range changes {
		path := strings.TrimSpace(change.Path)
		if path == "" {
			continue
		}
		lines = append(lines, traceFileChangeSymbol(change.Kind)+" "+path)
	}
	return strings.Join(lines, "\n")
}

func traceFileChangeSymbol(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "add", "added", "create", "created", "new":
		return "+"
	case "delete", "deleted", "remove", "removed":
		return "-"
	case "rename", "renamed", "move", "moved":
		return ">"
	case "modify", "modified", "update", "updated", "edit", "edited", "write", "wrote":
		return "~"
	default:
		return "*"
	}
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

var traceCodexToolNames = map[string]string{
	"command_execution": "Shell",
	"reasoning":         "Reasoning",
	"plan_update":       "Plan",
	"todo_list":         "Plan",
	"todo_write":        "Plan",
	"web_search":        "Web Search",
	"file_search":       "Search",
	"mcp_call":          "MCP Tool",
	"mcp_tool_call":     "MCP Tool",
	"file_change":       "File Change",
}

func traceCodexToolName(itemType string) string {
	if name := traceCodexToolNames[itemType]; name != "" {
		return name
	}
	return titleFromIdentifier(itemType)
}
