// Package claude implements a registry.Agent that shells out to the Claude
// Code CLI in headless mode and parses its stream-json NDJSON output via the
// agent/event/claude_types package.
package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/event/claude_types"
	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/exec"
)

// ClaudeAgent wraps the claude CLI (Anthropic's Claude Code agent).
type ClaudeAgent struct {
	AgentPath     string
	SettingsPath  string
	Workspace     string
	Env           *exec.Env
	LastSessionID string
}

// FindAgentPath looks up the claude binary in PATH.
func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("claude"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("claude not found in PATH")
}

// resolveAgentPath resolves the claude binary path via the registry resolver
// (considers agent path override, env var, settings, and PATH fallback).
func (a *ClaudeAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.ClaudeCLIPathSettingKey,
		registry.EnvClaudeCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("claude not found: %w", err)
	}
	return path, nil
}

// Ask invokes the claude CLI with the given prompt and streams the response.
//
// It runs:
//
//	claude -p "<prompt>" --output-format stream-json --verbose
//
// with optional --model and --resume flags. The --verbose flag is required
// by claude when using --output-format stream-json.
func (a *ClaudeAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}

	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}

	args := []string{
		"-p", question,
		"--output-format", "stream-json",
		"--verbose",
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	if opts != nil && opts.SessionID != "" {
		args = append(args, "--resume", opts.SessionID)
	}

	cmd := a.Env.CommandContext(ctx, agentPath, args...)
	cmd.Dir = workspace

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude: %w", err)
	}

	var eventWriter *ClaudeEventWriter
	if opts != nil && opts.RawLog != nil {
		eventWriter = NewClaudeEventWriter(opts.RawLog)
	}

	var fullAnswer strings.Builder
	var resultText string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if eventWriter != nil {
			eventWriter.WriteClaudeLine(line)
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "{") {
			continue
		}

		var ev claude_types.StreamEvent
		if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
			continue
		}

		if ev.SessionID != "" {
			a.LastSessionID = ev.SessionID
		}

		switch ev.Type {
		case claude_types.EventAssistant:
			if ev.Message != nil {
				for _, block := range ev.Message.Content {
					if block.Type == "text" && block.Text != "" {
						fullAnswer.WriteString(block.Text)
						if onDelta != nil {
							onDelta(block.Text)
						}
					}
				}
			}
		case claude_types.EventResult:
			if ev.Result != "" {
				resultText = ev.Result
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read claude output: %w", scanErr)
	}
	if eventWriter != nil {
		eventWriter.Flush()
	}

	if err := cmd.Wait(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("claude error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("claude exited with error: %w", err)
	}

	answer := fullAnswer.String()
	if answer == "" && resultText != "" {
		answer = resultText
	}
	return answer, nil
}

// ClaudeEventWriter converts claude stream-json NDJSON lines into coalesced
// AgentEvent JSONL. It is a thin pass-through to claude_types.FromClaude,
// which emits the correct 0..N canonical events per native line in order.
type ClaudeEventWriter struct {
	w io.Writer
}

// NewClaudeEventWriter creates a writer that emits AgentEvent JSONL to w.
func NewClaudeEventWriter(w io.Writer) *ClaudeEventWriter {
	return &ClaudeEventWriter{w: w}
}

// WriteClaudeLine parses one claude stream-json line and writes AgentEvent JSONL.
func (c *ClaudeEventWriter) WriteClaudeLine(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return
	}

	var ev claude_types.StreamEvent
	if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
		return
	}

	for _, agentEvent := range claude_types.FromClaude([]claude_types.StreamEvent{ev}, "") {
		writeAgentEvent(c.w, agentEvent)
	}
}

// Flush finalizes any buffered state. ClaudeEventWriter holds no pending
// state — FromClaude handles a single line's blocks in one call.
func (c *ClaudeEventWriter) Flush() {}

func writeAgentEvent(w io.Writer, event eventtypes.AgentEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n"))
}

// ListModels returns nil, nil because the claude CLI has no model-listing
// command.
func (a *ClaudeAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	return nil, nil
}
