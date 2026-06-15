package pi

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

type PiAgent struct {
	AgentPath     string
	SettingsPath  string
	Workspace     string
	Env           *exec.Env
	LastSessionID string
}

func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("pi"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("pi not found in PATH")
}

func (a *PiAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}

	args := []string{
		"--mode", "json",
		"--approve",
		"-p",
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts != nil && opts.SessionID != "" {
		args = append(args, "--session-id", opts.SessionID)
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
		return "", fmt.Errorf("failed to start pi: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	var streamError string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "{") {
			continue
		}

		var rawEvent map[string]any
		if err := json.Unmarshal([]byte(trimmed), &rawEvent); err != nil {
			continue
		}

		eventType, _ := rawEvent["type"].(string)
		if eventType == "" {
			continue
		}

		if eventType == "session" {
			if id, ok := rawEvent["id"].(string); ok && id != "" {
				a.LastSessionID = id
			}
			continue
		}

		switch eventType {
		case "agent_start", "turn_start":
		case "message_start":
			if msg, ok := rawEvent["message"].(map[string]any); ok {
				if sessionID, ok := msg["session_id"].(string); ok && sessionID != "" {
					a.LastSessionID = sessionID
				}
			}
		case "message_update":
			if msg, ok := rawEvent["message"].(map[string]any); ok {
				role, _ := msg["role"].(string)
				if role == "assistant" {
					content, _ := msg["content"].([]any)
					for _, c := range content {
						block, ok := c.(map[string]any)
						if !ok {
							continue
						}
						blockType, _ := block["type"].(string)
						if blockType == "text" {
							text, _ := block["text"].(string)
							if text != "" {
								fullAnswer.WriteString(text)
								if onDelta != nil {
									onDelta(text)
								}
							}
						}
					}
				}
			}
		case "message_end":
		case "tool_execution_start":
			if opts != nil && opts.OnToolCall != nil {
				toolName, _ := rawEvent["toolName"].(string)
				callID, _ := rawEvent["toolCallId"].(string)
				summary := ""
				if args, ok := rawEvent["args"]; ok {
					argsBytes, _ := json.Marshal(args)
					summary = truncateSummary(string(argsBytes))
				}
				opts.OnToolCall(registry.ToolCallEvent{
					Subtype:  "started",
					CallID:   callID,
					ToolName: toolName,
					Summary:  summary,
				})
			}
		case "tool_execution_end":
			if opts != nil && opts.OnToolCall != nil {
				toolName, _ := rawEvent["toolName"].(string)
				callID, _ := rawEvent["toolCallId"].(string)
				isError, _ := rawEvent["isError"].(bool)
				summary := ""
				if result, ok := rawEvent["result"]; ok {
					resultBytes, _ := json.Marshal(result)
					summary = truncateSummary(string(resultBytes))
				}
				status := "completed"
				if isError {
					status = "failed"
				}
				opts.OnToolCall(registry.ToolCallEvent{
					Subtype:  "completed",
					CallID:   callID,
					ToolName: toolName,
					Summary:  summary,
					Status:   status,
				})
			}
		case "turn_end":
			if msg, ok := rawEvent["message"].(map[string]any); ok {
				if sessionID, ok := msg["session_id"].(string); ok && sessionID != "" {
					a.LastSessionID = sessionID
				}
			}
		case "agent_end":
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read pi output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		if streamError != "" {
			return fullAnswer.String(), fmt.Errorf("%s", streamError)
		}
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("pi error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("pi exited with error: %w", err)
	}

	return fullAnswer.String(), nil
}

func truncateSummary(s string) string {
	if len(s) > 1000 {
		return s[:1000] + "..."
	}
	return s
}

func (a *PiAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	return nil, nil
}

func (a *PiAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.PiCLIPathSettingKey,
		registry.EnvPiCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("pi not found: %w", err)
	}
	return path, nil
}
