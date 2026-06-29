package traceparse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/agent/event/codex_types"
	"github.com/xhd2015/agent-pro/agent/event/summary"
	"github.com/xhd2015/agent-pro/agent/event/traceview"
)

type codexTraceAdapter struct{}

func init() {
	RegisterAdapter(10, codexTraceAdapter{})
}

func (codexTraceAdapter) Name() string {
	return "codex"
}

func (codexTraceAdapter) Parse(raw json.RawMessage) (traceview.AgentTraceParsedEvent, bool) {
	var event codex_types.TraceEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if event.Item == nil {
		return traceview.AgentTraceParsedEvent{}, false
	}
	if text := traceCodexAssistantText(event); text != "" {
		return traceview.AgentTraceParsedEvent{Message: &traceview.AgentTraceMessage{
			Role:    traceview.RoleAssistant,
			Content: text,
		}}, true
	}
	if activity := traceCodexActivity(event); activity != nil {
		return traceview.AgentTraceParsedEvent{Activity: activity}, true
	}
	return traceview.AgentTraceParsedEvent{}, false
}

func traceCodexAssistantText(event codex_types.TraceEvent) string {
	if event.Type != "item.completed" || event.Item == nil || !codex_types.TraceIsAssistantItem(event.Item.Type) {
		return ""
	}
	return codex_types.TraceItemText(event.Item)
}

const traceCodexHooksDeprecatedPrefix = "`[features].codex_hooks` is deprecated."

func traceCodexActivity(event codex_types.TraceEvent) *traceview.AgentTraceActivity {
	if event.Item == nil || codex_types.TraceIsAssistantItem(event.Item.Type) {
		return nil
	}
	var subtype traceview.ActivitySubtype
	switch event.Type {
	case "item.started":
		subtype = traceview.SubtypeStarted
	case "item.updated":
		subtype = traceview.SubtypeUpdated
	case "item.completed":
		subtype = traceview.SubtypeCompleted
	default:
		return nil
	}
	summaryText, replace := traceCodexSummary(event.Item, subtype)
	kind := event.Item.Type
	toolName := traceCodexToolName(kind)
	if isTraceCodexHooksDeprecation(summaryText) {
		kind = "warning"
		toolName = "Config Warning"
	}
	return &traceview.AgentTraceActivity{
		Subtype:        subtype,
		CallID:         event.Item.ID,
		ToolName:       toolName,
		Summary:        summaryText,
		Kind:           kind,
		Status:         traceCodexStatus(event.Item, subtype, summaryText),
		FileChanges:    event.Item.Changes,
		ReplaceSummary: replace,
	}
}

func traceCodexStatus(item *codex_types.TraceItem, subtype traceview.ActivitySubtype, summaryText string) traceview.ActivityStatus {
	if item == nil {
		return ""
	}
	if status := strings.TrimSpace(item.Status); status != "" {
		return traceview.ActivityStatus(status)
	}
	if item.Type == "error" {
		if isTraceCodexHooksDeprecation(summaryText) {
			return traceview.StatusWarning
		}
		return traceview.StatusFailed
	}
	if subtype == traceview.SubtypeCompleted {
		if item.ExitCode != nil && *item.ExitCode != 0 {
			return traceview.StatusFailed
		}
		return traceview.StatusCompleted
	}
	return traceview.StatusInProgress
}

func isTraceCodexHooksDeprecation(summaryText string) bool {
	return strings.HasPrefix(strings.TrimSpace(summaryText), traceCodexHooksDeprecatedPrefix)
}

func traceCodexSummary(item *codex_types.TraceItem, subtype traceview.ActivitySubtype) (string, bool) {
	switch item.Type {
	case "command_execution":
		if subtype == traceview.SubtypeStarted {
			return item.Command, false
		}
		return summarizeTraceCommandOutput(item), true
	case "reasoning":
		return codex_types.TraceItemText(item), subtype != traceview.SubtypeStarted
	case "todo_list", "todo_write":
		return summarizeTraceTodoList(item.Items), subtype != traceview.SubtypeStarted
	case "plan_update":
		return summarizeTracePlanUpdate(item), subtype != traceview.SubtypeStarted
	case "file_change":
		return summarizeTraceFileChanges(item.Changes), subtype != traceview.SubtypeStarted
	default:
		text := codex_types.TraceItemText(item)
		if text != "" {
			return text, subtype != traceview.SubtypeStarted
		}
		if item.Command != "" {
			return item.Command, subtype != traceview.SubtypeStarted
		}
		if item.Status != "" {
			return item.Status, subtype != traceview.SubtypeStarted
		}
		return "", subtype != traceview.SubtypeStarted
	}
}

func summarizeTraceCommandOutput(item *codex_types.TraceItem) string {
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
		b.WriteString(summary.CompactTraceOutput(output))
	}
	return b.String()
}

func summarizeTraceTodoList(items []codex_types.TraceTodoItem) string {
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

func summarizeTracePlanUpdate(item *codex_types.TraceItem) string {
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

func summarizeTraceFileChanges(changes []codex_types.FileChange) string {
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
	return summary.TitleFromIdentifier(itemType)
}