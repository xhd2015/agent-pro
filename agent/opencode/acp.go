package opencode

import (
	"context"
	"fmt"
	"os"
	"sync"

	agentacp "github.com/xhd2015/agent-pro/agent/acp"
)

type ACPAdapter struct {
	agent *OpencodeAgent

	mu       sync.Mutex
	sessions map[agentacp.SessionID]context.CancelFunc
}

var _ agentacp.Agent = (*ACPAdapter)(nil)

func NewACPAdapter(agent *OpencodeAgent) *ACPAdapter {
	return &ACPAdapter{
		agent:    agent,
		sessions: make(map[agentacp.SessionID]context.CancelFunc),
	}
}

func (a *ACPAdapter) Info() agentacp.AgentInfo {
	return agentacp.AgentInfo{
		Name:    "Opencode",
		Version: "1.0.0",
	}
}

func (a *ACPAdapter) Authenticate(req agentacp.AuthenticateRequest) error {
	return nil
}

func (a *ACPAdapter) NewSession(req agentacp.NewSessionRequest) (agentacp.SessionID, error) {
	cwd := req.Cwd
	if cwd == "" {
		cwd = a.agent.workspace()
	}

	opts := &SessionOpts{
		Dir:   cwd,
		Model: "",
	}

	ctx, cancel := context.WithCancel(context.Background())
	sessionID, err := a.agent.StartSession(ctx, "", opts, nil)
	if err != nil {
		cancel()
		return "", fmt.Errorf("new session: %w", err)
	}

	a.mu.Lock()
	a.sessions[agentacp.SessionID(sessionID)] = cancel
	a.mu.Unlock()

	return agentacp.SessionID(sessionID), nil
}

func (a *ACPAdapter) ResumeSession(req agentacp.ResumeSessionRequest) error {
	cwd := req.Cwd
	if cwd == "" {
		cwd = a.agent.workspace()
	}

	opts := &SessionOpts{
		Dir: cwd,
	}

	ctx, cancel := context.WithCancel(context.Background())
	_, err := a.agent.ResumeSession(ctx, string(req.SessionID), "", opts, nil)
	if err != nil {
		cancel()
		return fmt.Errorf("resume session: %w", err)
	}

	a.mu.Lock()
	a.sessions[req.SessionID] = cancel
	a.mu.Unlock()

	return nil
}

func (a *ACPAdapter) ListSessions(req agentacp.ListSessionsRequest) (agentacp.ListSessionsResponse, error) {
	ctx := context.Background()
	sessions, err := a.agent.ListSessions(ctx)
	if err != nil {
		return agentacp.ListSessionsResponse{}, fmt.Errorf("list sessions: %w", err)
	}

	infoList := make([]agentacp.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		infoList = append(infoList, agentacp.SessionInfo{
			SessionID: agentacp.SessionID(s.ID),
			Cwd:       s.Directory,
			Title:     s.Title,
			UpdatedAt: s.Updated.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return agentacp.ListSessionsResponse{Sessions: infoList}, nil
}

func (a *ACPAdapter) CloseSession(req agentacp.CloseSessionRequest) error {
	a.mu.Lock()
	cancel, ok := a.sessions[req.SessionID]
	delete(a.sessions, req.SessionID)
	a.mu.Unlock()

	if ok {
		cancel()
	}
	return nil
}

func (a *ACPAdapter) CancelSession(req agentacp.CancelSessionRequest) error {
	a.mu.Lock()
	cancel, ok := a.sessions[req.SessionID]
	a.mu.Unlock()

	if ok {
		cancel()
	}
	return nil
}

func (a *ACPAdapter) Prompt(req agentacp.PromptRequest, onUpdate func(agentacp.SessionUpdate)) error {
	exists := func() bool {
		a.mu.Lock()
		defer a.mu.Unlock()
		_, ok := a.sessions[req.SessionID]
		return ok
	}()

	if !exists {
		_, err := os.Stat(string(req.SessionID))
		if err != nil {
			return fmt.Errorf("unknown session: %s", req.SessionID)
		}
	}

	opts := &SessionOpts{}
	ctx, cancel := context.WithCancel(context.Background())

	if onUpdate == nil {
		onUpdate = func(agentacp.SessionUpdate) {}
	}

	_, _, err := a.agent.runSession(ctx, req.Content, string(req.SessionID), opts, func(event StreamEvent) {
		update := agentacp.SessionUpdate{}
		switch {
		case event.Text != "":
			update.Text = event.Text
		case event.ToolUse != nil:
			update.ToolCall = &agentacp.ToolCallUpdate{
				ID:      event.ToolUse.ID,
				Tool:    event.ToolUse.Tool,
				Status:  event.ToolUse.Status,
				Summary: event.ToolUse.Summary,
				Output:  event.ToolUse.Output,
				Error:   event.ToolUse.Error,
			}
		case event.Error != "":
			update.Error = event.Error
		case event.Done:
			update.Done = true
		}
		onUpdate(update)
	})
	if err != nil {
		cancel()
		return fmt.Errorf("prompt: %w", err)
	}

	cancel()
	return nil
}
