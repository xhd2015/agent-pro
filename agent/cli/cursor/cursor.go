package cursor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

type CursorAgent struct {
	AgentPath    string
	SettingsPath string
	Workspace    string
	Env          *exec.Env
}

func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("cursor-agent"); err == nil {
		return path, nil
	}
	if path, err := env.LookPath("agent"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("neither cursor-agent nor agent found in PATH")
}

func (a *CursorAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}
	sandboxMode := ""
	if opts != nil {
		sandboxMode = strings.TrimSpace(opts.SandboxMode)
	}
	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--stream-partial-output",
		"--trust",
		"--workspace", workspace,
	}
	if opts != nil && opts.AgentMode && sandboxMode == "" {
		args = append(args, "--yolo")
	} else {
		args = append(args, "--mode", "ask")
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	fullQuestion := buildAgentPrompt(workspace, question)
	if opts != nil && opts.DisableSubAgents {
		fullQuestion += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT use the Task tool (sub-agents/subagents) under any circumstances. Perform all work directly yourself without delegating to sub-agents."
	}
	args = append(args, fullQuestion)

	cmd := a.Env.CommandContext(ctx, agentPath, args...)
	cmd.Dir = workspace
	cmd.Env = a.Env.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	var stderrBuf strings.Builder
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start cursor-agent: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	gotResult := false
	hadToolCalls := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}

		var event agentEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Type {
		case "assistant":
			if event.TimestampMs == nil {
				continue
			}
			if event.ModelCallID != "" {
				continue
			}
			text := extractMessageText(event.Message)
			if text != "" {
				fullAnswer.WriteString(text)
				onDelta(text)
			}
		case "result":
			if event.Result != "" && fullAnswer.Len() == 0 {
				fullAnswer.WriteString(event.Result)
				onDelta(event.Result)
			}
		case "tool_call":
			if opts != nil && opts.OnToolCall != nil {
				toolName, summary := extractToolCallInfo(event.ToolCall, event.Subtype)
				opts.OnToolCall(registry.ToolCallEvent{
					Subtype:  event.Subtype,
					CallID:   event.CallID,
					ToolName: toolName,
					Summary:  summary,
				})
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		var agentErr error
		if stderrMsg != "" {
			agentErr = fmt.Errorf("cursor-agent error: %s", stderrMsg)
		} else {
			agentErr = fmt.Errorf("cursor-agent exited with error: %w", err)
		}
		return fullAnswer.String(), agentErr
	}

	if !gotResult && hadToolCalls {
		return fullAnswer.String(), fmt.Errorf("cursor-agent terminated without producing a final result (likely hit max tool call rounds)")
	}

	return fullAnswer.String(), nil
}

func (a *CursorAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return nil, err
	}
	cmd := a.Env.CommandContext(ctx, agentPath, "--list-models")
	cmd.Env = a.Env.Environ()

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	var models []registry.ModelInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Available") || strings.HasPrefix(line, "Tip:") || strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, " - ", 2)
		if len(parts) != 2 {
			continue
		}
		id := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		name = strings.TrimSuffix(name, "  (current)")
		name = strings.TrimSuffix(name, "  (default)")
		name = strings.TrimSpace(name)
		models = append(models, registry.ModelInfo{ID: id, Name: name})
	}
	return models, nil
}

func (a *CursorAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.CursorCLIPathSettingKey,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("cursor-agent not found: %w", err)
	}
	return path, nil
}

type agentEvent struct {
	Type        string                     `json:"type"`
	Subtype     string                     `json:"subtype,omitempty"`
	CallID      string                     `json:"call_id,omitempty"`
	ModelCallID string                     `json:"model_call_id,omitempty"`
	Message     *agentMsg                  `json:"message,omitempty"`
	Result      string                     `json:"result,omitempty"`
	TimestampMs *int64                     `json:"timestamp_ms,omitempty"`
	ToolCall    map[string]json.RawMessage `json:"tool_call,omitempty"`
}

type agentMsg struct {
	Role    string            `json:"role"`
	Content []agentMsgContent `json:"content"`
}

type agentMsgContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var friendlyToolNames = map[string]string{
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

func extractToolCallInfo(toolCall map[string]json.RawMessage, subtype string) (toolName string, summary string) {
	var rawKey string
	for key := range toolCall {
		rawKey = key
		break
	}

	toolName = rawKey
	if friendly, ok := friendlyToolNames[rawKey]; ok {
		toolName = friendly
	}

	raw := toolCall[rawKey]

	if subtype == "completed" {
		return toolName, extractCompletedSummary(rawKey, raw)
	}
	if subtype != "started" {
		return toolName, ""
	}
	var parsed struct {
		Args json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return toolName, ""
	}

	switch rawKey {
	case "updateTodosToolCall":
		return toolName, extractTodosSummary(parsed.Args)
	case "globToolCall":
		return toolName, extractGlobSummary(parsed.Args)
	case "taskToolCall":
		return toolName, extractTaskSummary(parsed.Args)
	}

	var args map[string]interface{}
	if err := json.Unmarshal(parsed.Args, &args); err != nil {
		return toolName, ""
	}
	if path, ok := args["path"].(string); ok {
		return toolName, path
	}
	if cmd, ok := args["command"].(string); ok {
		return toolName, cmd
	}
	if pattern, ok := args["pattern"].(string); ok {
		return toolName, pattern
	}
	if query, ok := args["query"].(string); ok {
		return toolName, query
	}
	return toolName, ""
}

func extractTodosSummary(argsRaw json.RawMessage) string {
	var args struct {
		Todos []struct {
			Content string `json:"content"`
			Status  string `json:"status"`
		} `json:"todos"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil || len(args.Todos) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, t := range args.Todos {
		if i > 0 {
			sb.WriteString("\n")
		}
		icon := "[ ]"
		switch t.Status {
		case "TODO_STATUS_IN_PROGRESS":
			icon = "[~]"
		case "TODO_STATUS_COMPLETED":
			icon = "[x]"
		}
		sb.WriteString(icon)
		sb.WriteString(" ")
		sb.WriteString(t.Content)
	}
	return sb.String()
}

func extractCompletedSummary(rawKey string, raw json.RawMessage) string {
	var parsed struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed.Result == nil {
		return ""
	}

	switch rawKey {
	case "shellToolCall":
		return extractShellResult(parsed.Result)
	case "taskToolCall":
		return extractTaskResult(parsed.Result)
	case "readToolCall":
		return extractReadResult(parsed.Result)
	case "globToolCall":
		return extractGlobResult(parsed.Result)
	case "grepToolCall":
		return extractGrepResult(parsed.Result)
	case "editToolCall", "writeToolCall":
		return extractGenericResult(parsed.Result)
	}
	return ""
}

func extractShellResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return ""
	}
	if _, ok := result["rejected"]; ok {
		return "[rejected]"
	}
	successRaw, ok := result["success"]
	if !ok {
		return ""
	}
	var success struct {
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exitCode"`
	}
	if err := json.Unmarshal(successRaw, &success); err != nil {
		return ""
	}
	output := strings.TrimSpace(success.Stdout)
	if success.ExitCode != 0 {
		return fmt.Sprintf("[exit %d]\n%s", success.ExitCode, output)
	}
	if output == "" {
		return "[done]"
	}
	return output
}

func extractTaskResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return "[done]"
	}
	successRaw, ok := result["success"]
	if !ok {
		return "[done]"
	}
	var success struct {
		DurationMs        string            `json:"durationMs"`
		ConversationSteps []json.RawMessage `json:"conversationSteps"`
	}
	if err := json.Unmarshal(successRaw, &success); err != nil {
		return "[completed]"
	}

	var header string
	if success.DurationMs != "" {
		var ms int
		if _, err := fmt.Sscanf(success.DurationMs, "%d", &ms); err == nil {
			header = fmt.Sprintf("[completed in %.1fs]", float64(ms)/1000)
		}
	}
	if header == "" {
		header = "[completed]"
	}

	responseText := extractLastAssistantMessage(success.ConversationSteps)
	if responseText == "" {
		return header
	}
	return header + "\n" + responseText
}

func extractLastAssistantMessage(steps []json.RawMessage) string {
	for i := len(steps) - 1; i >= 0; i-- {
		var step struct {
			AssistantMessage *struct {
				Text string `json:"text"`
			} `json:"assistantMessage"`
		}
		if err := json.Unmarshal(steps[i], &step); err != nil {
			continue
		}
		if step.AssistantMessage == nil || step.AssistantMessage.Text == "" {
			continue
		}
		return strings.TrimSpace(step.AssistantMessage.Text)
	}
	return ""
}

func extractReadResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return ""
	}
	if errRaw, ok := result["error"]; ok {
		var errMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.Unmarshal(errRaw, &errMsg); err == nil && errMsg.ErrorMessage != "" {
			return "[error: " + errMsg.ErrorMessage + "]"
		}
		return "[error]"
	}
	successRaw, ok := result["success"]
	if !ok {
		return ""
	}
	var success struct {
		TotalLines    int  `json:"totalLines"`
		FileSize      int  `json:"fileSize"`
		IsEmpty       bool `json:"isEmpty"`
		ExceededLimit bool `json:"exceededLimit"`
	}
	if err := json.Unmarshal(successRaw, &success); err != nil {
		return "[read ok]"
	}
	if success.IsEmpty {
		return "[empty file]"
	}
	if success.ExceededLimit {
		return fmt.Sprintf("[read %d lines, %d bytes, truncated]", success.TotalLines, success.FileSize)
	}
	return fmt.Sprintf("[read %d lines, %d bytes]", success.TotalLines, success.FileSize)
}

func extractGlobResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return ""
	}
	successRaw, ok := result["success"]
	if !ok {
		return "[no matches]"
	}
	var success struct {
		Files      []string `json:"files"`
		TotalFiles int      `json:"totalFiles"`
	}
	if err := json.Unmarshal(successRaw, &success); err != nil {
		return ""
	}
	if success.TotalFiles == 0 {
		return "[no matches]"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%d file(s)]\n", success.TotalFiles))
	for i, f := range success.Files {
		if i >= 20 {
			sb.WriteString(fmt.Sprintf("... and %d more", success.TotalFiles-20))
			break
		}
		sb.WriteString(f)
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func extractGrepResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return ""
	}
	if errRaw, ok := result["error"]; ok {
		var errMsg struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(errRaw, &errMsg); err == nil && errMsg.Error != "" {
			return "[error: " + errMsg.Error + "]"
		}
		return "[error]"
	}
	successRaw, ok := result["success"]
	if !ok {
		return "[no matches]"
	}
	var success struct {
		OutputMode string `json:"outputMode"`
		Pattern    string `json:"pattern"`
	}
	if err := json.Unmarshal(successRaw, &success); err != nil {
		return "[grep done]"
	}
	return "[matches found]"
}

func extractGenericResult(resultRaw json.RawMessage) string {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return ""
	}
	if _, ok := result["error"]; ok {
		return "[error]"
	}
	if _, ok := result["rejected"]; ok {
		return "[rejected]"
	}
	return "[done]"
}

func extractTaskSummary(argsRaw json.RawMessage) string {
	var args struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return ""
	}
	return args.Description
}

func extractGlobSummary(argsRaw json.RawMessage) string {
	var args struct {
		Pattern         string `json:"pattern"`
		GlobPattern     string `json:"globPattern"`
		Path            string `json:"path"`
		TargetDirectory string `json:"targetDirectory"`
	}
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return ""
	}
	pattern := args.GlobPattern
	if pattern == "" {
		pattern = args.Pattern
	}
	dir := args.TargetDirectory
	if dir == "" {
		dir = args.Path
	}
	if pattern == "" {
		return dir
	}
	if dir != "" {
		return dir + "/" + pattern
	}
	return pattern
}

func buildAgentPrompt(workspace string, question string) string {
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

func extractMessageText(msg *agentMsg) string {
	if msg == nil {
		return ""
	}
	var sb strings.Builder
	for _, c := range msg.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return sb.String()
}
