package grok

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/agent/exec"
)

// GrokAgent wraps the grok CLI (xAI's Grok coding agent).
type GrokAgent struct {
	AgentPath     string
	SettingsPath  string
	Workspace     string
	Env           *exec.Env
	LastSessionID string
}

// FindAgentPath looks up the grok binary in PATH.
func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("grok"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("grok not found in PATH")
}

// resolveAgentPath resolves the grok binary path via the registry resolver
// (considers agent path override, env var, settings, and PATH fallback).
func (a *GrokAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.GrokCLIPathSettingKey,
		registry.EnvGrokCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("grok not found: %w", err)
	}
	return path, nil
}

// Ask invokes the grok CLI with the given prompt and streams the response.
//
// It runs:
//
//	grok -p "<prompt>" --output-format streaming-json --always-approve
//
// with optional --model and --resume flags.
func (a *GrokAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
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
		"--output-format", "streaming-json",
		"--always-approve",
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
		return "", fmt.Errorf("failed to start grok: %w", err)
	}

	var eventWriter *GrokEventWriter
	if opts != nil && opts.RawLog != nil {
		eventWriter = NewGrokEventWriter(opts.RawLog)
	}

	var fullAnswer strings.Builder
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if eventWriter != nil {
			eventWriter.WriteGrokLine(line)
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "{") {
			continue
		}

		var ev struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
			continue
		}

		switch ev.Type {
		case "thought":
			// discard
		case "text":
			if ev.Data != "" {
				fullAnswer.WriteString(ev.Data)
				if onDelta != nil {
					onDelta(ev.Data)
				}
			}
		case "end":
			// The "end" event carries stopReason, sessionId, requestId
			var endEv struct {
				Type      string `json:"type"`
				SessionID string `json:"sessionId"`
			}
			if err := json.Unmarshal([]byte(trimmed), &endEv); err == nil {
				if endEv.SessionID != "" {
					a.LastSessionID = endEv.SessionID
				}
			}
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read grok output: %w", scanErr)
	}
	if eventWriter != nil {
		eventWriter.Flush()
	}

	if err := cmd.Wait(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("grok error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("grok exited with error: %w", err)
	}

	return fullAnswer.String(), nil
}

// GrokEventWriter converts grok streaming JSON lines into coalesced AgentEvent JSONL.
type GrokEventWriter struct {
	w            io.Writer
	pendingThink strings.Builder
}

// NewGrokEventWriter creates a writer that buffers consecutive grok thought deltas
// into a single ActionThink event before writing to w.
func NewGrokEventWriter(w io.Writer) *GrokEventWriter {
	return &GrokEventWriter{w: w}
}

// WriteGrokLine parses one grok streaming JSON line and writes AgentEvent JSONL.
func (g *GrokEventWriter) WriteGrokLine(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") {
		return
	}

	var ev grok_types.Event
	if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
		return
	}

	if ev.Type == grok_types.EventThought {
		if strings.TrimSpace(ev.Data) == "" {
			return
		}
		g.pendingThink.WriteString(ev.Data)
		return
	}

	g.flushThink()

	switch ev.Type {
	case grok_types.EventText:
		if strings.TrimSpace(ev.Data) == "" {
			return
		}
	}

	for _, agentEvent := range grok_types.FromGrok([]grok_types.Event{ev}) {
		writeAgentEvent(g.w, agentEvent)
	}
}

// Flush writes any buffered thought text as a single ActionThink event.
func (g *GrokEventWriter) Flush() {
	g.flushThink()
}

func (g *GrokEventWriter) flushThink() {
	if g.pendingThink.Len() == 0 {
		return
	}
	writeAgentEvent(g.w, eventtypes.AgentEvent{
		Type: eventtypes.ActionThink,
		Text: g.pendingThink.String(),
	})
	g.pendingThink.Reset()
}

func writeAgentEventsFromGrokLine(rawLog io.Writer, line string) {
	if rawLog == nil {
		return
	}
	w := NewGrokEventWriter(rawLog)
	w.WriteGrokLine(line)
	w.Flush()
}

func writeAgentEvent(w io.Writer, event eventtypes.AgentEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n"))
}

// ListModels runs "grok models" and parses the output to return model IDs.
//
// Sample output:
//
//	You are logged in with grok.com.
//
//	Default model: grok-composer-2.5-fast
//
//	Available models:
//	  - grok-build
//	  * grok-composer-2.5-fast (default)
//
// Lines starting with "- " or "* " are parsed as model entries.
func (a *GrokAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return nil, err
	}

	cmd := a.Env.CommandContext(ctx, agentPath, "models")
	cmd.Dir = a.Workspace

	var stdoutBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("grok models error: %s", stderrMsg)
		}
		return nil, fmt.Errorf("grok models exited with error: %w", err)
	}

	output := stdoutBuf.String()
	return parseModelsOutput(output), nil
}

// parseModelsOutput extracts model IDs from the plain-text output of "grok models".
// It scans lines starting with "- " or "* " and extracts the model ID token.
func parseModelsOutput(output string) []registry.ModelInfo {
	var models []registry.ModelInfo
	seen := make(map[string]bool)

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Lines starting with "* " or "- " indicate a model entry.
		if !strings.HasPrefix(trimmed, "* ") && !strings.HasPrefix(trimmed, "- ") {
			continue
		}

		// Strip the prefix and extract model ID (first space-delimited token).
		rest := trimmed[2:]
		// Strip potential " (default)" suffix
		rest = strings.TrimSuffix(rest, " (default)")
		modelID := strings.Fields(rest)[0]
		if modelID != "" && !seen[modelID] {
			seen[modelID] = true
			models = append(models, registry.ModelInfo{ID: modelID, Name: modelID})
		}
	}

	return models
}
