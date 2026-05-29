package codex

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/agent_trace/types"
)

type codexTraceAdapter struct{}

func init() {
	types.RegisterAgentTraceAdapter(10, codexTraceAdapter{})
}

func (codexTraceAdapter) Name() string {
	return "codex"
}

func (codexTraceAdapter) Parse(raw json.RawMessage) (types.AgentTraceParsedEvent, bool) {
	var event types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return types.AgentTraceParsedEvent{}, false
	}
	if event.Item == nil {
		return types.AgentTraceParsedEvent{}, false
	}
	if text := traceCodexAssistantText(event); text != "" {
		return types.AgentTraceParsedEvent{Message: &types.AgentTraceMessage{
			Role:    types.RoleAssistant,
			Content: text,
		}}, true
	}
	if activity := traceCodexActivity(event); activity != nil {
		return types.AgentTraceParsedEvent{Activity: activity}, true
	}
	return types.AgentTraceParsedEvent{}, false
}

func traceCodexAssistantText(event types.TraceEvent) string {
	if event.Type != "item.completed" || event.Item == nil || !types.TraceIsAssistantItem(event.Item.Type) {
		return ""
	}
	return types.TraceItemText(event.Item)
}

const traceCodexHooksDeprecatedPrefix = "`[features].codex_hooks` is deprecated."

func traceCodexActivity(event types.TraceEvent) *types.AgentTraceActivity {
	if event.Item == nil || types.TraceIsAssistantItem(event.Item.Type) {
		return nil
	}
	var subtype types.ActivitySubtype
	switch event.Type {
	case "item.started":
		subtype = types.SubtypeStarted
	case "item.updated":
		subtype = types.SubtypeUpdated
	case "item.completed":
		subtype = types.SubtypeCompleted
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
	return &types.AgentTraceActivity{
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

func traceCodexStatus(item *types.TraceItem, subtype types.ActivitySubtype, summary string) types.ActivityStatus {
	if item == nil {
		return ""
	}
	if status := strings.TrimSpace(item.Status); status != "" {
		return types.ActivityStatus(status)
	}
	if item.Type == "error" {
		if isTraceCodexHooksDeprecation(summary) {
			return types.StatusWarning
		}
		return types.StatusFailed
	}
	if subtype == types.SubtypeCompleted {
		if item.ExitCode != nil && *item.ExitCode != 0 {
			return types.StatusFailed
		}
		return types.StatusCompleted
	}
	return types.StatusInProgress
}

func isTraceCodexHooksDeprecation(summary string) bool {
	return strings.HasPrefix(strings.TrimSpace(summary), traceCodexHooksDeprecatedPrefix)
}

func traceCodexSummary(item *types.TraceItem, subtype types.ActivitySubtype) (string, bool) {
	switch item.Type {
	case "command_execution":
		if subtype == types.SubtypeStarted {
			return item.Command, false
		}
		return summarizeTraceCommandOutput(item), true
	case "reasoning":
		return types.TraceItemText(item), subtype != types.SubtypeStarted
	case "todo_list", "todo_write":
		return summarizeTraceTodoList(item.Items), subtype != types.SubtypeStarted
	case "plan_update":
		return summarizeTracePlanUpdate(item), subtype != types.SubtypeStarted
	case "file_change":
		return summarizeTraceFileChanges(item.Changes), subtype != types.SubtypeStarted
	default:
		text := types.TraceItemText(item)
		if text != "" {
			return text, subtype != types.SubtypeStarted
		}
		if item.Command != "" {
			return item.Command, subtype != types.SubtypeStarted
		}
		if item.Status != "" {
			return item.Status, subtype != types.SubtypeStarted
		}
		return "", subtype != types.SubtypeStarted
	}
}

func summarizeTraceCommandOutput(item *types.TraceItem) string {
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
		b.WriteString(types.CompactTraceOutput(output))
	}
	return b.String()
}

func summarizeTraceTodoList(items []types.TraceTodoItem) string {
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

func summarizeTracePlanUpdate(item *types.TraceItem) string {
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

func summarizeTraceFileChanges(changes []types.FileChange) string {
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
	return types.TitleFromIdentifier(itemType)
}
