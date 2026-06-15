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
	"github.com/xhd2015/agent-pro/agent/event/pi_types"
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

		var rawEvent pi_types.Event
		if err := json.Unmarshal([]byte(trimmed), &rawEvent); err != nil {
			continue
		}

		if rawEvent.Type == "" {
			continue
		}

		if rawEvent.Type == pi_types.EventTypeSession {
			if rawEvent.ID != "" {
				a.LastSessionID = rawEvent.ID
			}
			continue
		}

		switch rawEvent.Type {
		case pi_types.EventTypeAgentStart, pi_types.EventTypeTurnStart:
		case pi_types.EventTypeMessageStart:
			if rawEvent.Message != nil && rawEvent.Message.SessionID != "" {
				a.LastSessionID = rawEvent.Message.SessionID
			}
		case pi_types.EventTypeMessageUpdate:
			if rawEvent.Message != nil && rawEvent.Message.Role == "assistant" {
				for _, c := range rawEvent.Message.Content {
					if c.Type == "text" && c.Text != "" {
						fullAnswer.WriteString(c.Text)
						if onDelta != nil {
							onDelta(c.Text)
						}
					}
				}
			}
		case pi_types.EventTypeMessageEnd:
		case pi_types.EventTypeToolExecStart:
			if opts != nil && opts.OnToolCall != nil {
				summary := ""
				if rawEvent.Args != nil {
					argsBytes, _ := json.Marshal(rawEvent.Args)
					summary = truncateSummary(string(argsBytes))
				}
				opts.OnToolCall(registry.ToolCallEvent{
					Subtype:  "started",
					CallID:   rawEvent.ToolCallID,
					ToolName: rawEvent.ToolName,
					Summary:  summary,
				})
			}
		case pi_types.EventTypeToolExecEnd:
			if opts != nil && opts.OnToolCall != nil {
				summary := ""
				if rawEvent.Result != nil {
					resultBytes, _ := json.Marshal(rawEvent.Result)
					summary = truncateSummary(string(resultBytes))
				}
				status := "completed"
				if rawEvent.IsError {
					status = "failed"
				}
				opts.OnToolCall(registry.ToolCallEvent{
					Subtype:  "completed",
					CallID:   rawEvent.ToolCallID,
					ToolName: rawEvent.ToolName,
					Summary:  summary,
					Status:   status,
				})
			}
		case pi_types.EventTypeTurnEnd:
			if rawEvent.Message != nil && rawEvent.Message.SessionID != "" {
				a.LastSessionID = rawEvent.Message.SessionID
			}
		case pi_types.EventTypeAgentEnd:
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
