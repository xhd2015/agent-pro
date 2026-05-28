package opencode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type OpencodeAgent struct {
	AgentPath     string
	Workspace     string
	Env           []string
	DisableSubAgents bool
}

type AgentOption func(*OpencodeAgent)

func WithAgentPath(path string) AgentOption {
	return func(o *OpencodeAgent) {
		o.AgentPath = path
	}
}

func WithWorkspace(dir string) AgentOption {
	return func(o *OpencodeAgent) {
		o.Workspace = dir
	}
}

func WithEnv(env []string) AgentOption {
	return func(o *OpencodeAgent) {
		o.Env = env
	}
}

func WithDisableSubAgents() AgentOption {
	return func(o *OpencodeAgent) {
		o.DisableSubAgents = true
	}
}

func NewAgent(opts ...AgentOption) *OpencodeAgent {
	a := &OpencodeAgent{}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func FindAgentPath() (string, error) {
	path, err := exec.LookPath("opencode")
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("opencode not found in PATH")
}

func (a *OpencodeAgent) resolvePath() (string, error) {
	if a.AgentPath != "" {
		if _, err := os.Stat(a.AgentPath); err == nil {
			return a.AgentPath, nil
		}
	}
	return FindAgentPath()
}

func (a *OpencodeAgent) workspace() string {
	if a.Workspace != "" {
		return a.Workspace
	}
	wd, _ := os.Getwd()
	return wd
}

func (a *OpencodeAgent) ListModels(ctx context.Context) ([]Model, error) {
	agentPath, err := a.resolvePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, agentPath, "models")
	cmd.Env = a.buildEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}

	models := []Model{}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		model := Model{
			ID:   line,
			Name: line,
		}
		if parts := strings.SplitN(line, "/", 2); len(parts) == 2 {
			model.ProviderID = parts[0]
			model.ID = parts[1]
		}
		models = append(models, model)
	}
	return models, nil
}

func (a *OpencodeAgent) ListSessions(ctx context.Context) ([]Session, error) {
	agentPath, err := a.resolvePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, agentPath, "session", "list", "--format", "json")
	cmd.Env = a.buildEnv()
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	var raw []struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Updated   int64  `json:"updated"`
		Created   int64  `json:"created"`
		ProjectID string `json:"projectId,omitempty"`
		Directory string `json:"directory,omitempty"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}

	sessions := make([]Session, 0, len(raw))
	for _, s := range raw {
		sessions = append(sessions, Session{
			ID:        s.ID,
			Title:     s.Title,
			Created:   msToTime(s.Created),
			Updated:   msToTime(s.Updated),
			ProjectID: s.ProjectID,
			Directory: s.Directory,
		})
	}
	return sessions, nil
}

func (a *OpencodeAgent) StartSession(ctx context.Context, prompt string, opts *SessionOpts, onEvent StreamCallback) (string, error) {
	return a.runSession(ctx, prompt, "", opts, onEvent)
}

func (a *OpencodeAgent) ResumeSession(ctx context.Context, sessionID string, prompt string, opts *SessionOpts, onEvent StreamCallback) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("sessionID is required for resume")
	}
	return a.runSession(ctx, prompt, sessionID, opts, onEvent)
}

func (a *OpencodeAgent) runSession(ctx context.Context, prompt string, sessionID string, opts *SessionOpts, onEvent StreamCallback) (string, error) {
	agentPath, err := a.resolvePath()
	if err != nil {
		return "", err
	}

	args := []string{
		"run",
		"--format", "json",
	}
	if sessionID != "" {
		args = append(args, "--session", sessionID)
	}
	args = append(args, "--dir", a.workspace())
	args = append(args, "--dangerously-skip-permissions")

	if opts != nil {
		if opts.Model != "" {
			model := opts.Model
			if !strings.Contains(model, "/") {
				model = "opencode/" + model
			}
			args = append(args, "--model", model)
		}
		if opts.Agent != "" {
			args = append(args, "--agent", opts.Agent)
		}
		if opts.Dir != "" {
			args = append(args, "--dir", opts.Dir)
		}
		if opts.Command != "" {
			args = append(args, "--command", opts.Command)
		}
		if opts.Variant != "" {
			args = append(args, "--variant", opts.Variant)
		}
		for _, f := range opts.File {
			args = append(args, "--file", f)
		}
	}

	if a.DisableSubAgents || (opts != nil && opts.NoSubAgents) {
		prompt += "\n\n# CRITICAL RULE: DO NOT USE SUB-AGENTS\nYou MUST NOT use the Task tool (sub-agents/subagents) under any circumstances. Perform all work directly yourself without delegating to sub-agents."
	}

	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, agentPath, args...)
	cmd.Dir = a.workspace()
	cmd.Env = a.buildEnv()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start opencode: %w", err)
	}

	sessionIDResult := sessionID
	fullAnswer := strings.Builder{}

	done := make(chan error, 2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "{") {
				continue
			}

			event := parseStreamEvent(line, onEvent)
			if event.SessionID != "" && sessionIDResult == "" {
				sessionIDResult = event.SessionID
			}
			if event.Text != "" {
				fullAnswer.WriteString(event.Text)
			}
		}
		done <- scanner.Err()
	}()
	go func() {
		stderrBuf := strings.Builder{}
		io.Copy(&stderrBuf, stderr)
		if stderrBuf.Len() > 0 {
			done <- fmt.Errorf("%s", strings.TrimSpace(stderrBuf.String()))
		} else {
			done <- nil
		}
	}()

	var firstErr error
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	waitErr := cmd.Wait()
	if firstErr != nil {
		return sessionIDResult, firstErr
	}
	if waitErr != nil {
		return sessionIDResult, fmt.Errorf("opencode exited: %w", waitErr)
	}

	if onEvent != nil {
		onEvent(StreamEvent{Type: "done", Done: true})
	}
	return sessionIDResult, nil
}

func (a *OpencodeAgent) buildEnv() []string {
	if len(a.Env) > 0 {
		return a.Env
	}
	return os.Environ()
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

func parseStreamEvent(line string, onEvent StreamCallback) StreamEvent {
	var raw struct {
		Type      string          `json:"type"`
		Timestamp int64           `json:"timestamp,omitempty"`
		SessionID string          `json:"sessionID,omitempty"`
		Part      json.RawMessage `json:"part,omitempty"`
		Error     json.RawMessage `json:"error,omitempty"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return StreamEvent{}
	}

	event := StreamEvent{
		Type:      raw.Type,
		Timestamp: raw.Timestamp,
		SessionID: raw.SessionID,
	}

	switch raw.Type {
	case "text":
		var part struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			event.Text = part.Text
		}
	case "tool_use":
		var part struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Tool   string `json:"tool"`
			CallID string `json:"callID,omitempty"`
			State  *struct {
				Status string `json:"status"`
				Title  string `json:"title"`
				Output string `json:"output"`
				Error  string `json:"error"`
			} `json:"state,omitempty"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			toolUse := &ToolUseEvent{
				ID:   part.ID,
				Tool: part.Tool,
			}
			if part.State != nil {
				toolUse.Status = part.State.Status
				toolUse.Summary = part.State.Title
				toolUse.Output = part.State.Output
				toolUse.Error = part.State.Error
			}
			if toolUse.Status == "" {
				toolUse.Status = "started"
			}
			event.ToolUse = toolUse
		}
	case "error":
		var errData struct {
			Name string `json:"name"`
			Data struct {
				Message string `json:"message"`
			} `json:"data"`
		}
		if json.Unmarshal(raw.Error, &errData) == nil && errData.Name != "" {
			event.Error = errData.Name
			if errData.Data.Message != "" {
				event.Error = errData.Data.Message
			}
		}
		if event.Error == "" {
			event.Error = "opencode error"
		}
	case "reasoning":
		var part struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(raw.Part, &part) == nil {
			event.Reasoning = part.Text
		}
	}

	if onEvent != nil {
		onEvent(event)
	}

	return event
}
