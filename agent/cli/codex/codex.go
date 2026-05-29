package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

type CodexAgent struct {
	AgentPath    string
	SettingsPath string
	Workspace    string
	Env          *exec.Env
}

func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("codex"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("codex not found in PATH")
}

func (a *CodexAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}
	sandboxMode := "danger-full-access"
	if opts != nil && strings.TrimSpace(opts.SandboxMode) != "" {
		sandboxMode = strings.TrimSpace(opts.SandboxMode)
	}

	args := []string{
		"exec",
		"--json",
		"--skip-git-repo-check",
		"--cd", workspace,
		"--sandbox", sandboxMode,
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	fullQuestion := buildPrompt(workspace, question)
	if opts != nil && opts.DisableSubAgents {
		fullQuestion += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT delegate work to sub-agents or spawn parallel agents. Perform all work directly yourself."
	}
	args = append(args, fullQuestion)

	cmd := a.Env.CommandContext(ctx, agentPath, args...)
	cmd.Dir = workspace
	cmd.Env = a.commandEnv()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	var stderrBuf strings.Builder
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start codex: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	reportedStreamErrors := make(map[string]bool)
	latestStreamError := ""
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "{") {
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if streamErr := event.extractStreamError(); streamErr != nil {
			latestStreamError = streamErr.FinalMessage
			if opts != nil && opts.OnToolCall != nil && streamErr.ToolEvent != nil {
				key := streamErr.ToolEvent.CallID + "\n" + streamErr.ToolEvent.Summary
				if !reportedStreamErrors[key] {
					reportedStreamErrors[key] = true
					opts.OnToolCall(*streamErr.ToolEvent)
				}
			}
			continue
		}

		if opts != nil && opts.OnToolCall != nil {
			if toolEvent := event.extractToolCallEvent(); toolEvent != nil {
				opts.OnToolCall(*toolEvent)
			}
		}

		if text := event.extractAssistantText(); text != "" {
			fullAnswer.WriteString(text)
			onDelta(text)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read codex output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		if latestStreamError != "" {
			return fullAnswer.String(), fmt.Errorf("%s", latestStreamError)
		}
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("codex error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("codex exited with error: %w", err)
	}
	return fullAnswer.String(), nil
}

func (a *CodexAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return nil, err
	}
	models, err := a.listModelsFromCatalog(ctx, agentPath)
	if err != nil {
		return []registry.ModelInfo{
			{ID: "", Name: "Default"},
		}, nil
	}
	return models, nil
}

type codexModelCatalog struct {
	Models []codexModelEntry `json:"models"`
}

type codexModelEntry struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Visibility  string `json:"visibility"`
}

func (a *CodexAgent) listModelsFromCatalog(ctx context.Context, agentPath string) ([]registry.ModelInfo, error) {
	cmd := a.Env.CommandContext(ctx, agentPath, "debug", "models", "--bundled")
	cmd.Env = a.commandEnv()

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list codex models: %w", err)
	}

	var catalog codexModelCatalog
	if err := json.Unmarshal(out, &catalog); err != nil {
		return nil, fmt.Errorf("parse codex model catalog: %w", err)
	}

	models := []registry.ModelInfo{{ID: "", Name: "Default"}}
	seen := map[string]bool{"": true}
	for _, model := range catalog.Models {
		id := strings.TrimSpace(model.Slug)
		if id == "" || seen[id] {
			continue
		}
		if visibility := strings.TrimSpace(model.Visibility); visibility != "" && visibility != "list" {
			continue
		}
		name := strings.TrimSpace(model.DisplayName)
		if name == "" {
			name = id
		}
		models = append(models, registry.ModelInfo{
			ID:   id,
			Name: name,
		})
		seen[id] = true
	}
	return models, nil
}

func (a *CodexAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.CodexCLIPathSettingKey,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("codex not found: %w", err)
	}
	return path, nil
}

func (a *CodexAgent) commandEnv() []string {
	env := a.Env.Environ()
	apiKey := registry.LoadConfiguredStringSetting(a.SettingsPath, registry.CodexAPIKeySettingKey)
	if apiKey == "" {
		return env
	}
	return upsertEnv(env, "OPENAI_API_KEY", apiKey)
}

func upsertEnv(env []string, key string, value string) []string {
	prefix := key + "="
	updated := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			if !replaced {
				updated = append(updated, prefix+value)
				replaced = true
			}
			continue
		}
		updated = append(updated, entry)
	}
	if !replaced {
		updated = append(updated, prefix+value)
	}
	return updated
}

type codexEvent struct {
	Type    string     `json:"type"`
	Item    *codexItem `json:"item,omitempty"`
	Delta   string     `json:"delta,omitempty"`
	Text    string     `json:"text,omitempty"`
	Message string     `json:"message,omitempty"`
}

type codexItem struct {
	ID               string          `json:"id,omitempty"`
	Type             string          `json:"type,omitempty"`
	Text             string          `json:"text,omitempty"`
	Content          []codexItemPart `json:"content,omitempty"`
	Command          string          `json:"command,omitempty"`
	AggregatedOutput string          `json:"aggregated_output,omitempty"`
	ExitCode         *int            `json:"exit_code,omitempty"`
	Status           string          `json:"status,omitempty"`
	Raw              map[string]any  `json:"-"`
}

type codexItemPart struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

func (i *codexItem) UnmarshalJSON(data []byte) error {
	type alias codexItem
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*i = codexItem(parsed)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		i.Raw = raw
	}
	return nil
}

func (e codexEvent) extractAssistantText() string {
	switch e.Type {
	case "item.completed":
		return extractItemText(e.Item)
	default:
		return ""
	}
}

func (e codexEvent) extractToolCallEvent() *registry.ToolCallEvent {
	if e.Item == nil || shouldIgnoreCodexItem(e.Item.Type) {
		return nil
	}
	subtype := mapCodexEventSubtype(e.Type)
	if subtype == "" {
		return nil
	}
	fileChanges := extractFileChanges(e.Item.Raw)
	summary, replaceSummary := summarizeCodexItem(e.Item, subtype, fileChanges)
	event := &registry.ToolCallEvent{
		Subtype:        subtype,
		CallID:         e.Item.ID,
		ToolName:       friendlyCodexItemName(e.Item.Type),
		Summary:        summary,
		Kind:           e.Item.Type,
		Status:         codexItemStatus(e.Item, subtype),
		FileChanges:    fileChanges,
		ReplaceSummary: replaceSummary,
	}
	if event.Status == "" && subtype == "completed" {
		event.Status = "completed"
	}
	if event.Status == "" && (subtype == "started" || subtype == "updated") {
		event.Status = "in_progress"
	}
	return event
}

type codexStreamError struct {
	ToolEvent    *registry.ToolCallEvent
	FinalMessage string
}

func (e codexEvent) extractStreamError() *codexStreamError {
	if e.Type != "error" {
		return nil
	}
	message := strings.TrimSpace(firstNonEmpty(e.Message, e.Text))
	if message == "" {
		return nil
	}
	normalized := normalizeCodexErrorMessage(message)
	toolName := "Codex"
	kind := "runtime_error"
	callID := "codex_runtime_error"
	if normalized.IsAuth {
		toolName = "Authentication"
		kind = "auth_error"
		callID = "codex_auth_error"
	}
	return &codexStreamError{
		ToolEvent: &registry.ToolCallEvent{
			Subtype:        "completed",
			CallID:         callID,
			ToolName:       toolName,
			Summary:        normalized.DisplayMessage,
			Kind:           kind,
			Status:         "failed",
			ReplaceSummary: true,
		},
		FinalMessage: normalized.FinalMessage,
	}
}

type normalizedCodexError struct {
	DisplayMessage string
	FinalMessage   string
	IsAuth         bool
}

func normalizeCodexErrorMessage(message string) normalizedCodexError {
	trimmed := strings.TrimSpace(message)
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "401 unauthorized") || strings.Contains(lower, "missing bearer or basic authentication") {
		display := "Codex CLI is not authenticated.\nOpenAI request failed with 401 Unauthorized: missing bearer/basic authentication.\nSign in to Codex CLI or provide `OPENAI_API_KEY` for the knowledge-portal process."
		return normalizedCodexError{
			DisplayMessage: display,
			FinalMessage:   "codex authentication failed: OpenAI request returned 401 Unauthorized because no authentication header was provided; sign in to Codex CLI or provide OPENAI_API_KEY",
			IsAuth:         true,
		}
	}
	if strings.Contains(lower, "reconnecting...") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "Reconnecting..."))
	}
	return normalizedCodexError{
		DisplayMessage: trimmed,
		FinalMessage:   "codex error: " + trimmed,
	}
}

func mapCodexEventSubtype(eventType string) string {
	switch eventType {
	case "item.started":
		return "started"
	case "item.updated":
		return "updated"
	case "item.completed":
		return "completed"
	default:
		return ""
	}
}

func shouldIgnoreCodexItem(itemType string) bool {
	switch itemType {
	case "", "agent_message", "message", "assistant_message", "output_text":
		return true
	default:
		return false
	}
}

var codexFriendlyItemNames = map[string]string{
	"command_execution": "Shell",
	"reasoning":         "Reasoning",
	"plan_update":       "Plan",
	"todo_list":         "Plan",
	"todo_write":        "Plan",
	"web_search":        "Web Search",
	"file_search":       "Search",
	"mcp_call":          "MCP Tool",
	"mcp_tool_call":     "MCP Tool",
}

func friendlyCodexItemName(itemType string) string {
	if itemType == "" {
		return "Codex Event"
	}
	if friendly, ok := codexFriendlyItemNames[itemType]; ok {
		return friendly
	}
	parts := strings.FieldsFunc(itemType, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	if len(parts) == 0 {
		return itemType
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func extractItemText(item *codexItem) string {
	if item == nil {
		return ""
	}
	switch item.Type {
	case "agent_message", "message", "assistant_message", "output_text":
	default:
		if item.Text == "" && len(item.Content) == 0 {
			return ""
		}
	}
	if item.Text != "" {
		return item.Text
	}
	var sb strings.Builder
	for _, part := range item.Content {
		if part.Text != "" {
			sb.WriteString(part.Text)
		}
	}
	return sb.String()
}

func summarizeCodexItem(item *codexItem, subtype string, fileChanges []registry.FileChange) (summary string, replaceSummary bool) {
	if item == nil {
		return "", false
	}
	switch item.Type {
	case "command_execution":
		return summarizeCommandExecution(item, subtype), subtype != "started"
	case "file_change":
		return summarizeFileChanges(fileChanges, subtype), true
	case "mcp_call", "mcp_tool_call":
		return summarizeMCPToolCall(item.Raw, subtype), subtype != "started"
	case "reasoning":
		if text := firstNonEmpty(
			extractItemText(item),
			extractRawString(item.Raw, "summary"),
			extractRawString(item.Raw, "text"),
			extractRawStrings(item.Raw, "content", "steps", "entries"),
		); text != "" {
			return text, subtype != "started"
		}
		if subtype == "started" {
			return "Thinking", false
		}
		return "", false
	default:
		if text := firstNonEmpty(
			extractItemText(item),
			extractRawString(item.Raw, "summary"),
			extractRawString(item.Raw, "title"),
			extractRawString(item.Raw, "message"),
			extractRawString(item.Raw, "query"),
			extractRawString(item.Raw, "path"),
			extractRawString(item.Raw, "url"),
			extractRawStrings(item.Raw, "todos", "steps", "entries", "items", "results"),
		); text != "" {
			return text, subtype != "started"
		}
		if subtype == "started" {
			switch item.Type {
			case "todo_list", "todo_write", "plan_update":
				return "Updating plan", false
			}
		}
		return "", false
	}
}

func summarizeMCPToolCall(raw map[string]any, subtype string) string {
	if raw == nil {
		return ""
	}

	var lines []string
	nameParts := []string{}
	if server := extractRawString(raw, "server"); server != "" {
		nameParts = append(nameParts, server)
	}
	if tool := extractRawString(raw, "tool"); tool != "" {
		nameParts = append(nameParts, tool)
	}
	if len(nameParts) > 0 {
		lines = append(lines, strings.Join(nameParts, "."))
	}

	if args := summarizeMCPArguments(raw["arguments"]); args != "" {
		lines = append(lines, args)
	}

	if subtype != "started" {
		if result := firstNonEmpty(
			extractRawString(raw, "error"),
			extractRawString(raw, "result"),
		); result != "" {
			lines = append(lines, result)
		}
	}

	return strings.Join(lines, "\n")
}

func summarizeMCPArguments(value any) string {
	args, ok := value.(map[string]any)
	if !ok || len(args) == 0 {
		return ""
	}
	for _, key := range []string{"doc", "url", "path", "query", "q"} {
		if text := flattenRawValue(args[key]); text != "" {
			return key + ": " + text
		}
	}
	data, err := json.Marshal(args)
	if err != nil {
		return ""
	}
	return "arguments: " + string(data)
}

func summarizeFileChanges(changes []registry.FileChange, subtype string) string {
	if len(changes) == 0 {
		if subtype == "started" {
			return "Tracking file changes"
		}
		return ""
	}

	lines := []string{fmt.Sprintf("%d %s changed", len(changes), pluralize(len(changes), "file", "files"))}
	for _, change := range changes {
		path := strings.TrimSpace(change.Path)
		if path == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s", fileChangeSymbol(change.Kind), path))
	}
	return strings.Join(lines, "\n")
}

func summarizeCommandExecution(item *codexItem, subtype string) string {
	if item == nil {
		return ""
	}

	command := strings.TrimSpace(item.Command)
	output := strings.TrimSpace(item.AggregatedOutput)
	exitLabel := ""
	if item.ExitCode != nil {
		switch *item.ExitCode {
		case 0:
			exitLabel = "[done]"
		default:
			exitLabel = fmt.Sprintf("[exit %d]", *item.ExitCode)
		}
	}

	if subtype == "started" {
		return command
	}
	var lines []string
	if command != "" {
		lines = append(lines, command)
	}
	switch {
	case output == "" && exitLabel == "":
		return strings.Join(lines, "\n")
	case output == "":
		lines = append(lines, exitLabel)
	case exitLabel == "":
		lines = append(lines, output)
	default:
		lines = append(lines, exitLabel, output)
	}
	return strings.Join(lines, "\n")
}

func codexItemStatus(item *codexItem, subtype string) string {
	if item == nil {
		return ""
	}
	if item.ExitCode != nil && *item.ExitCode != 0 {
		return "failed"
	}
	if item.Status != "" {
		return item.Status
	}
	switch subtype {
	case "started", "updated":
		return "in_progress"
	case "completed":
		return "completed"
	default:
		return ""
	}
}

func extractFileChanges(raw map[string]any) []registry.FileChange {
	if raw == nil {
		return nil
	}
	values, ok := raw["changes"].([]any)
	if !ok || len(values) == 0 {
		return nil
	}
	changes := make([]registry.FileChange, 0, len(values))
	for _, value := range values {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		path := strings.TrimSpace(flattenRawValue(entry["path"]))
		if path == "" {
			continue
		}
		changes = append(changes, registry.FileChange{
			Path: filepath.Clean(path),
			Kind: normalizeFileChangeKind(flattenRawValue(entry["kind"])),
		})
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

func normalizeFileChangeKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "add", "added", "create", "created", "new":
		return "add"
	case "delete", "deleted", "remove", "removed":
		return "delete"
	case "rename", "renamed", "move", "moved":
		return "rename"
	case "modify", "modified", "update", "updated", "edit", "edited", "write", "wrote":
		return "modify"
	default:
		return strings.ToLower(strings.TrimSpace(kind))
	}
}

func fileChangeSymbol(kind string) string {
	switch normalizeFileChangeKind(kind) {
	case "add":
		return "+"
	case "delete":
		return "-"
	case "rename":
		return ">"
	case "modify":
		return "~"
	default:
		return "*"
	}
}

func extractRawString(raw map[string]any, key string) string {
	if raw == nil {
		return ""
	}
	value, ok := raw[key]
	if !ok {
		return ""
	}
	return flattenRawValue(value)
}

func extractRawStrings(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if text := extractRawString(raw, key); text != "" {
			return text
		}
	}
	return ""
}

func flattenRawValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		var parts []string
		for _, item := range v {
			if text := flattenRawValue(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	case map[string]any:
		for _, key := range []string{"text", "content", "summary", "title", "description", "message", "query", "path", "url"} {
			if text := flattenRawValue(v[key]); text != "" {
				return text
			}
		}
		return ""
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func pluralize(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func buildPrompt(workspace string, question string) string {
	taskPath := workspace + "/TASK.md"
	tmplPath := workspace + "/OUTPUT_TEMPLATE.md"

	taskContent, taskErr := os.ReadFile(taskPath)
	tmplContent, tmplErr := os.ReadFile(tmplPath)

	hasTask := taskErr == nil && len(taskContent) > 0
	hasTmpl := tmplErr == nil && len(tmplContent) > 0

	if !hasTask && !hasTmpl {
		return question
	}

	var sb strings.Builder
	if hasTask {
		sb.WriteString("# Task Definition (from TASK.md)\n\n")
		sb.WriteString(string(taskContent))
		sb.WriteString("\n\n")
	}
	if hasTmpl {
		sb.WriteString("# Output Template (from OUTPUT_TEMPLATE.md)\n\n")
		sb.WriteString(string(tmplContent))
		sb.WriteString("\n\n")
	}
	sb.WriteString("# User Question\n\n")
	sb.WriteString(question)
	return sb.String()
}
