package commandcode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/agent/exec"
)

// CommandcodeAgent wraps the Command Code CLI (`cmd`).
type CommandcodeAgent struct {
	AgentPath     string
	SettingsPath  string
	Workspace     string
	Env           *exec.Env
	LastSessionID string
}

const defaultMaxTurns = "32"

// FindAgentPath looks up the Command Code binary (`cmd`, then `commandcode`).
func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("cmd"); err == nil {
		return path, nil
	}
	if path, err := env.LookPath("commandcode"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("commandcode (cmd) not found in PATH")
}

func (a *CommandcodeAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.CommandcodeCLIPathSettingKey,
		registry.EnvCommandcodeCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("commandcode not found: %w", err)
	}
	return path, nil
}

// Ask runs `cmd -p` with JSON output and optional --resume / --model.
func (a *CommandcodeAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return "", err
	}

	workspace := a.Workspace
	if opts != nil && opts.Workspace != "" {
		workspace = opts.Workspace
	}

	fullQuestion := question
	if opts != nil && opts.DisableSubAgents {
		fullQuestion += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT delegate work to sub-agents. Perform all work directly yourself."
	}

	args := []string{
		"-p", fullQuestion,
		"--skip-onboarding",
		"--yolo",
		"--max-turns", defaultMaxTurns,
		"--output-format", "json",
	}
	if opts != nil && opts.Model != "" {
		args = append(args, "-m", opts.Model)
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
		return "", fmt.Errorf("failed to start commandcode: %w", err)
	}

	rawLog := io.Writer(nil)
	if opts != nil {
		rawLog = opts.RawLog
	}

	var fullAnswer strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
			continue
		}

		text, sessionID := parseCommandcodeJSONLine(trimmed)
		if sessionID != "" {
			a.LastSessionID = sessionID
		}
		if text != "" {
			fullAnswer.WriteString(text)
			if onDelta != nil {
				onDelta(text)
			}
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read commandcode output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("commandcode error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("commandcode exited with error: %w", err)
	}
	return fullAnswer.String(), nil
}

func (a *CommandcodeAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	return []registry.ModelInfo{{ID: "", Name: "Default"}}, nil
}

// parseCommandcodeJSONLine extracts assistant text and session id from one NDJSON line.
func parseCommandcodeJSONLine(line string) (text string, sessionID string) {
	var top struct {
		Type      string          `json:"type"`
		SessionID string          `json:"sessionId"`
		FinalText string          `json:"finalText"`
		Event     json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal([]byte(line), &top); err != nil {
		return "", ""
	}
	sessionID = strings.TrimSpace(top.SessionID)

	switch top.Type {
	case "result":
		return strings.TrimSpace(top.FinalText), sessionID
	case "event":
		if len(top.Event) == 0 {
			return "", sessionID
		}
		var ev struct {
			Type      string `json:"type"`
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(top.Event, &ev); err != nil {
			return "", sessionID
		}
		if sid := strings.TrimSpace(ev.SessionID); sid != "" {
			sessionID = sid
		}
		return "", sessionID
	default:
		return "", sessionID
	}
}
