package opencode

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

type OpencodeAgent struct {
	AgentPath    string
	SettingsPath string
	Workspace    string
	Env          *exec.Env
}

func FindAgentPath(env *exec.Env) (string, error) {
	if path, err := env.LookPath("opencode"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("opencode not found in PATH")
}

func (a *OpencodeAgent) Ask(ctx context.Context, question string, opts *registry.AskOptions, onDelta registry.DeltaCallback) (string, error) {
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
		"--format", "json",
		"--dir", workspace,
		"--dangerously-skip-permissions",
	}
	if opts != nil && opts.Model != "" {
		model := opts.Model
		if !strings.Contains(model, "/") {
			model = "openai/" + model
		}
		args = append(args, "--model", model)
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
		return "", fmt.Errorf("failed to start opencode: %w", err)
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
		if strings.TrimSpace(line) == "" {
			continue
		}
		if rawLog != nil {
			_, _ = rawLog.Write([]byte(line + "\n"))
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "{") {
			continue
		}

		var event opencodeRunEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Type {
		case "text":
			if event.Part != nil {
				text := strings.TrimSpace(event.Part.Text)
				if text != "" {
					fullAnswer.WriteString(text)
					onDelta(text)
				}
			}
		case "tool_use":
			if opts != nil && opts.OnToolCall != nil && event.Part != nil {
				toolEvent := mapOpencodeToolEvent(event.Part)
				if toolEvent != nil {
					opts.OnToolCall(*toolEvent)
				}
			}
		case "error":
			errMsg := ""
			if event.Error != nil {
				errMsg, _ = event.Error["name"].(string)
				if data, ok := event.Error["data"].(map[string]any); ok {
					if msg, ok := data["message"].(string); ok && msg != "" {
						errMsg = msg
					}
				}
			}
			if errMsg == "" {
				errMsg = "opencode error"
			}
			stderrBuf.WriteString(errMsg + "\n")
		case "reasoning":
		case "step_start":
		case "step_finish":
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fullAnswer.String(), fmt.Errorf("failed to read opencode output: %w", scanErr)
	}

	if err := cmd.Wait(); err != nil {
		stderrMsg := strings.TrimSpace(stderrBuf.String())
		if stderrMsg != "" {
			return fullAnswer.String(), fmt.Errorf("opencode error: %s", stderrMsg)
		}
		return fullAnswer.String(), fmt.Errorf("opencode exited with error: %w", err)
	}
	return fullAnswer.String(), nil
}

func (a *OpencodeAgent) ListModels(ctx context.Context) ([]registry.ModelInfo, error) {
	agentPath, err := a.resolveAgentPath()
	if err != nil {
		return nil, err
	}

	cmd := a.Env.CommandContext(ctx, agentPath, "models")
	cmd.Env = a.Env.Environ()

	out, err := cmd.Output()
	if err != nil {
		return []registry.ModelInfo{
			{ID: "", Name: "Default"},
		}, nil
	}

	models := []registry.ModelInfo{{ID: "", Name: "Default"}}
	seen := map[string]bool{"": true}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if modelsByName, ok := parseOpencodeModel(line); ok {
			if !seen[modelsByName.ID] {
				models = append(models, modelsByName)
				seen[modelsByName.ID] = true
			}
		}
	}
	return models, nil
}

func (a *OpencodeAgent) resolveAgentPath() (string, error) {
	path, err := registry.ResolveConfiguredCLIPath(
		a.SettingsPath,
		registry.OpencodeCLIPathSettingKey,
		registry.EnvOpencodeCLIPath,
		a.AgentPath,
		func() (string, error) { return FindAgentPath(a.Env) },
	)
	if err != nil {
		return "", fmt.Errorf("opencode not found: %w", err)
	}
	return path, nil
}

type opencodeRunEvent struct {
	Type      string                `json:"type"`
	Timestamp int64                 `json:"timestamp,omitempty"`
	SessionID string                `json:"sessionID,omitempty"`
	Part      *opencodeRunEventPart `json:"part,omitempty"`
	Error     map[string]any        `json:"error,omitempty"`
}

type opencodeRunEventPart struct {
	ID     string             `json:"id"`
	Type   string             `json:"type"`
	Tool   string             `json:"tool,omitempty"`
	Text   string             `json:"text,omitempty"`
	CallID string             `json:"callID,omitempty"`
	State  *opencodePartState `json:"state,omitempty"`
}

type opencodePartState struct {
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
	Title  string         `json:"title,omitempty"`
	Output string         `json:"output,omitempty"`
	Input  map[string]any `json:"input,omitempty"`
}

func mapOpencodeToolEvent(part *opencodeRunEventPart) *registry.ToolCallEvent {
	if part == nil {
		return nil
	}
	subtype := "started"
	status := "in_progress"
	summary := ""
	output := ""

	if part.State != nil {
		switch part.State.Status {
		case "pending":
			subtype = "started"
			status = "pending"
		case "running":
			subtype = "started"
			status = "in_progress"
		case "completed":
			subtype = "completed"
			status = "completed"
			summary = part.State.Title
			output = part.State.Output
			if output != "" {
				if summary != "" {
					summary = summary + "\n" + output
				} else {
					summary = output
				}
			}
		case "error":
			subtype = "completed"
			status = "failed"
			summary = part.State.Error
		}

		if summary == "" && part.State.Title != "" {
			summary = part.State.Title
		}
		if summary == "" && part.State.Input != nil {
			if cmd, ok := part.State.Input["command"].(string); ok {
				summary = cmd
			} else if path, ok := part.State.Input["path"].(string); ok {
				summary = path
			} else if pattern, ok := part.State.Input["pattern"].(string); ok {
				summary = pattern
			} else if query, ok := part.State.Input["query"].(string); ok {
				summary = query
			}
		}
	}

	toolName := part.Tool
	if toolName == "" {
		toolName = part.Type
	}
	friendlyName := friendlyOpencodeToolName(toolName)

	return &registry.ToolCallEvent{
		Subtype:  subtype,
		CallID:   part.CallID,
		ToolName: friendlyName,
		Summary:  summary,
		Kind:     part.Tool,
		Status:   status,
	}
}

var opencodeFriendlyToolNames = map[string]string{
	"bash":      "Shell",
	"Read":      "Read File",
	"read":      "Read File",
	"Edit":      "Edit File",
	"edit":      "Edit File",
	"Write":     "Write File",
	"write":     "Write File",
	"Glob":      "Glob",
	"glob":      "Glob",
	"Grep":      "Grep",
	"grep":      "Grep",
	"WebSearch": "Web Search",
	"websearch": "Web Search",
	"WebFetch":  "Web Fetch",
	"webfetch":  "Web Fetch",
	"Task":      "Sub-Agent",
	"task":      "Sub-Agent",
	"TodoWrite": "Plan",
	"todowrite": "Plan",
	"LSP":       "LSP",
	"lsp":       "LSP",
	"Skill":     "Skill",
	"skill":     "Skill",
}

func friendlyOpencodeToolName(tool string) string {
	if tool == "" {
		return "Tool"
	}
	if friendly, ok := opencodeFriendlyToolNames[tool]; ok {
		return friendly
	}
	return tool
}

func parseOpencodeModel(line string) (registry.ModelInfo, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return registry.ModelInfo{}, false
	}
	return registry.ModelInfo{
		ID:   line,
		Name: line,
	}, true
}
