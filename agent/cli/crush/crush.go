package crush

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
	crush_types "github.com/xhd2015/agent-pro/agent/event/crush_types"
)

type CrushAgent struct {
	AgentPath     string
	SettingsPath  string
	Workspace     string
	Env           *exec.Env
	LastSessionID string
	ServerClient  *CrushServerClient
}

func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("crush"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("crush not found in PATH")
}

func (a *CrushAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	if a.ServerClient != nil {
		return a.askViaServer(ctx, question, opts, onDelta)
	}
	return a.askViaSubprocess(ctx, question, opts, onDelta)
}

func (a *CrushAgent) askViaSubprocess(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}

	args := []string{
		"run",
		"--verbose",
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts != nil && opts.SessionID != "" {
		args = append(args, "--session", opts.SessionID)
	}
	fullQuestion := question
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
		return "", fmt.Errorf("failed to start crush: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	var streamError string
	var plainText strings.Builder
	hasEvent := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}

		data := ExtractSSEData(line)
		if data == "" {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !isLogLine(trimmed) {
				plainText.WriteString(trimmed)
				plainText.WriteString("\n")
			}
			continue
		}

		var event crush_types.Event
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		hasEvent = true
		switch event.Type {
		case crush_types.EventMessage:
			a.processMessageEvent(event.Payload, opts, onDelta, &fullAnswer, &streamError)
		case crush_types.EventAgentEvent:
			a.processAgentEvent(event.Payload, &streamError)
		case crush_types.EventRunComplete:
			a.processRunComplete(event.Payload, &fullAnswer, &streamError)
		}
	}

	if !hasEvent && fullAnswer.Len() == 0 && plainText.Len() > 0 {
		answer := strings.TrimSpace(plainText.String())
		fullAnswer.WriteString(answer)
		if onDelta != nil {
			onDelta(answer)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read crush output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		if streamError != "" {
			return fullAnswer.String(), fmt.Errorf("%s", streamError)
		}
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("crush error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("crush exited with error: %w", err)
	}

	a.extractSessionIDFromStderr(stderrBuf.String())

	return fullAnswer.String(), nil
}

func (a *CrushAgent) askViaServer(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	client := a.ServerClient

	fullQuestion := question
	if opts != nil && opts.DisableSubAgents {
		fullQuestion += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT use the Task tool (sub-agents/subagents) under any circumstances. Perform all work directly yourself without delegating to sub-agents."
	}

	if err := client.EnsureServer(ctx); err != nil {
		return "", fmt.Errorf("ensure server: %w", err)
	}

	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}
	if workspace == "" {
		workspace = "."
	}
	workspaceID, err := client.CreateWorkspace(ctx, workspace)
	if err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}

	if err := client.InitAgent(ctx, workspaceID); err != nil {
		return "", fmt.Errorf("init agent: %w", err)
	}

	eventCh, err := client.SubscribeEvents(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("subscribe events: %w", err)
	}

	sessionID := ""
	if opts != nil {
		sessionID = opts.SessionID
	}
	if sessionID == "" {
		sessionID, err = client.CreateSession(ctx, workspaceID, "")
		if err != nil {
			return "", fmt.Errorf("create session: %w", err)
		}
	}
	if err := client.SendMessage(ctx, workspaceID, fullQuestion, sessionID); err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	var streamError string
eventLoop:
	for dataLine := range eventCh {
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(dataLine + "\n"))
		}

		data := ExtractSSEData(dataLine)
		if data == "" {
			continue
		}

		event, err := UnwrapEvent([]byte(data))
		if err != nil {
			continue
		}
		if event == nil {
			continue
		}

		switch event.Type {
		case crush_types.EventMessage:
			a.processMessageEvent(event.Payload, opts, onDelta, &fullAnswer, &streamError)
		case crush_types.EventAgentEvent:
			a.processAgentEvent(event.Payload, &streamError)
		case crush_types.EventRunComplete:
			a.processRunComplete(event.Payload, &fullAnswer, &streamError)
			break eventLoop
		}
	}

	if streamError != "" {
		return fullAnswer.String(), fmt.Errorf("%s", streamError)
	}

	return fullAnswer.String(), nil
}

func (a *CrushAgent) extractSessionIDFromStderr(stderr string) {
	if stderr == "" {
		return
	}
	const prefix = "session_id="
	idx := strings.Index(stderr, prefix)
	if idx < 0 {
		return
	}
	start := idx + len(prefix)
	end := start
	for end < len(stderr) {
		c := stderr[end]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			end++
		} else {
			break
		}
	}
	sessionID := stderr[start:end]
	if sessionID != "" {
		a.LastSessionID = sessionID
	}
}

func (a *CrushAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	return nil, nil
}

func (a *CrushAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.CrushCLIPathSettingKey,
		registry.EnvCrushCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("crush not found: %w", err)
	}
	return path, nil
}

func (a *CrushAgent) processMessageEvent(payload json.RawMessage, opts *registry.AskOptions, onDelta registry.DeltaCallback, fullAnswer *strings.Builder, streamError *string) {
	var msg crush_types.MessagePayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	if msg.SessionID != "" {
		a.LastSessionID = msg.SessionID
	}
	if msg.Role != "assistant" {
		return
	}
	for _, part := range msg.Parts {
		switch part.Type {
		case crush_types.PartReasoning:
			d := partReasoningData(part)
			if d != nil && d.Thinking != "" {
				fullAnswer.WriteString(d.Thinking)
			}
		case crush_types.PartText:
			d := partTextData(part)
			if d != nil {
				text := strings.TrimSpace(d.Text)
				if text != "" {
					fullAnswer.WriteString(text)
					if onDelta != nil {
						onDelta(text)
					}
				}
			}
		case crush_types.PartToolCall:
			if opts != nil && opts.OnToolCall != nil {
				d := partToolCallData(part)
				if d != nil {
					opts.OnToolCall(registry.ToolCallEvent{
						Subtype:  "started",
						CallID:   d.ID,
						ToolName: d.Name,
						Summary:  d.Input,
					})
				}
			}
		case crush_types.PartToolResult:
			if opts != nil && opts.OnToolCall != nil {
				var result struct {
					ToolCallID string `json:"tool_call_id"`
					Name       string `json:"name"`
					Content    string `json:"content"`
				}
				if err := json.Unmarshal(part.Data, &result); err == nil {
					opts.OnToolCall(registry.ToolCallEvent{
						Subtype:  "completed",
						CallID:   result.ToolCallID,
						ToolName: result.Name,
						Summary:  result.Content,
					})
				}
			}
		case crush_types.PartFinish:
			d := partFinishData(part)
			if d != nil && d.Reason == crush_types.FinishReasonError {
				if *streamError == "" {
					*streamError = "crush stream error: " + d.Message
				}
			}
		}
	}
}

func (a *CrushAgent) processAgentEvent(payload json.RawMessage, streamError *string) {
	var evt crush_types.AgentEventPayload
	if err := json.Unmarshal(payload, &evt); err != nil {
		return
	}
	if evt.Type == "error" && *streamError == "" {
		*streamError = "crush error: " + evt.Error
	}
}

func (a *CrushAgent) processRunComplete(payload json.RawMessage, fullAnswer *strings.Builder, streamError *string) {
	var rc crush_types.RunCompletePayload
	if err := json.Unmarshal(payload, &rc); err != nil {
		return
	}
	if rc.SessionID != "" {
		a.LastSessionID = rc.SessionID
	}
	if rc.Text != "" {
		fullAnswer.WriteString(rc.Text)
	}
	if rc.Error != "" && *streamError == "" {
		*streamError = rc.Error
	}
}

func ExtractSSEData(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	const dataPrefix = "data:"
	data := trimmed
	if strings.HasPrefix(trimmed, dataPrefix) {
		data = strings.TrimSpace(strings.TrimPrefix(trimmed, dataPrefix))
	}
	if data == "" || !strings.HasPrefix(data, "{") {
		return ""
	}
	return data
}

func isLogLine(line string) bool {
	for _, prefix := range []string{"INFO", "ERRO", "WARN", "DEBU"} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func partReasoningData(p crush_types.Part) *crush_types.ReasoningData {
	var d crush_types.ReasoningData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func partTextData(p crush_types.Part) *crush_types.TextData {
	var d crush_types.TextData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func partToolCallData(p crush_types.Part) *crush_types.ToolCallData {
	var d crush_types.ToolCallData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}

func partFinishData(p crush_types.Part) *crush_types.FinishData {
	var d crush_types.FinishData
	if err := json.Unmarshal(p.Data, &d); err != nil {
		return nil
	}
	return &d
}
