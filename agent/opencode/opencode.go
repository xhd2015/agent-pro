package opencode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/opencode/run"
)

type OpencodeAgent struct {
	AgentPath        string
	Workspace        string
	Env              []string
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

func (a *OpencodeAgent) ResolvePath() (string, error) {
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
	agentPath, err := a.ResolvePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, agentPath, "models")
	cmd.Env = a.BuildEnv()
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
			model.AgentRunnerID = parts[0]
			model.ID = parts[1]
		}
		models = append(models, model)
	}
	return models, nil
}

func (a *OpencodeAgent) ListSessions(ctx context.Context) ([]Session, error) {
	agentPath, err := a.ResolvePath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, agentPath, "session", "list", "--format", "json")
	cmd.Env = a.BuildEnv()
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
	sessionID, _, err := a.runSession(ctx, prompt, "", opts, onEvent)
	return sessionID, err
}

func (a *OpencodeAgent) ResumeSession(ctx context.Context, sessionID string, prompt string, opts *SessionOpts, onEvent StreamCallback) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("sessionID is required for resume")
	}
	sessionID, _, err := a.runSession(ctx, prompt, sessionID, opts, onEvent)
	return sessionID, err
}

func (a *OpencodeAgent) runSession(ctx context.Context, prompt string, sessionID string, opts *SessionOpts, onEvent StreamCallback) (string, string, error) {
	agentPath, err := a.ResolvePath()
	if err != nil {
		return "", "", err
	}

	sessionOpts := run.SessionRunOpts{
		AgentPath:        agentPath,
		Dir:              a.workspace(),
		Env:              a.BuildEnv(),
		SessionID:        sessionID,
		Prompt:           prompt,
		OnEvent:          run.StreamCallback(onEvent),
		DisableSubAgents: a.DisableSubAgents,
	}
	if opts != nil {
		sessionOpts.Model = opts.Model
		sessionOpts.Agent = opts.Agent
		sessionOpts.Dir = opts.Dir
		sessionOpts.File = opts.File
		sessionOpts.Command = opts.Command
		sessionOpts.Variant = opts.Variant
		sessionOpts.NoSubAgents = opts.NoSubAgents
		sessionOpts.Thinking = opts.Thinking
	}

	sessionID, fullAnswer, err := run.StartSession(ctx, sessionOpts)
	return sessionID, fullAnswer, err
}

func (a *OpencodeAgent) BuildEnv() []string {
	if len(a.Env) > 0 {
		return a.Env
	}
	return os.Environ()
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}
