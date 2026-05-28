package acp

import (
	"context"

	acpsdk "github.com/coder/acp-go-sdk"
)

func (b *agentBridge) Authenticate(ctx context.Context, params acpsdk.AuthenticateRequest) (acpsdk.AuthenticateResponse, error) {
	return acpsdk.AuthenticateResponse{}, nil
}

func (b *agentBridge) Initialize(ctx context.Context, params acpsdk.InitializeRequest) (acpsdk.InitializeResponse, error) {
	info := b.agent.Info()
	return acpsdk.InitializeResponse{
		ProtocolVersion: acpsdk.ProtocolVersionNumber,
		AgentInfo: &acpsdk.Implementation{
			Name:    info.Name,
			Version: info.Version,
		},
		AgentCapabilities: acpsdk.AgentCapabilities{
			SessionCapabilities: acpsdk.SessionCapabilities{
				List:   &acpsdk.SessionListCapabilities{},
				Resume: &acpsdk.SessionResumeCapabilities{},
				Close:  &acpsdk.SessionCloseCapabilities{},
			},
		},
	}, nil
}

func (b *agentBridge) Cancel(ctx context.Context, params acpsdk.CancelNotification) error {
	return b.agent.CancelSession(CancelSessionRequest{
		SessionID: SessionID(params.SessionId),
	})
}

func (b *agentBridge) CloseSession(ctx context.Context, params acpsdk.CloseSessionRequest) (acpsdk.CloseSessionResponse, error) {
	err := b.agent.CloseSession(CloseSessionRequest{
		SessionID: SessionID(params.SessionId),
	})
	return acpsdk.CloseSessionResponse{}, err
}

func (b *agentBridge) ListSessions(ctx context.Context, params acpsdk.ListSessionsRequest) (acpsdk.ListSessionsResponse, error) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	cwd := ""
	if params.Cwd != nil {
		cwd = *params.Cwd
	}
	resp, err := b.agent.ListSessions(ListSessionsRequest{
		Cwd:    cwd,
		Cursor: cursor,
	})
	if err != nil {
		return acpsdk.ListSessionsResponse{}, err
	}
	sdkSessions := make([]acpsdk.SessionInfo, len(resp.Sessions))
	for i, s := range resp.Sessions {
		title := s.Title
		sdkSessions[i] = acpsdk.SessionInfo{
			SessionId: acpsdk.SessionId(s.SessionID),
			Cwd:       s.Cwd,
			Title:     &title,
			UpdatedAt: &s.UpdatedAt,
		}
	}
	var nextCursor *string
	if resp.NextCursor != "" {
		nextCursor = &resp.NextCursor
	}
	return acpsdk.ListSessionsResponse{
		Sessions:   sdkSessions,
		NextCursor: nextCursor,
	}, nil
}

func (b *agentBridge) NewSession(ctx context.Context, params acpsdk.NewSessionRequest) (acpsdk.NewSessionResponse, error) {
	meta := params.Meta
	sessionID, err := b.agent.NewSession(NewSessionRequest{
		Cwd:  params.Cwd,
		Meta: meta,
	})
	if err != nil {
		return acpsdk.NewSessionResponse{}, err
	}
	return acpsdk.NewSessionResponse{
		SessionId: acpsdk.SessionId(sessionID),
	}, nil
}

func (b *agentBridge) Prompt(ctx context.Context, params acpsdk.PromptRequest) (acpsdk.PromptResponse, error) {
	req := PromptRequest{
		SessionID: SessionID(params.SessionId),
		Meta:      params.Meta,
	}
	for _, block := range params.Prompt {
		if block.Text != nil {
			req.Content += block.Text.Text
		}
	}

	err := b.agent.Prompt(req, func(update SessionUpdate) {
		b.sendUpdate(ctx, params.SessionId, update)
	})
	if err != nil {
		return acpsdk.PromptResponse{}, err
	}
	return acpsdk.PromptResponse{
		StopReason: acpsdk.StopReasonEndTurn,
	}, nil
}

func (b *agentBridge) sendUpdate(ctx context.Context, sessionID acpsdk.SessionId, update SessionUpdate) {
	conn := b.conn
	if conn == nil {
		return
	}
	switch {
	case update.Text != "":
		conn.SessionUpdate(ctx, acpsdk.SessionNotification{
			SessionId: sessionID,
			Update:    acpsdk.UpdateAgentMessageText(update.Text),
		})
	case update.ToolCall != nil:
		tc := update.ToolCall
		conn.SessionUpdate(ctx, acpsdk.SessionNotification{
			SessionId: sessionID,
			Update:    acpsdk.StartToolCall(acpsdk.ToolCallId(tc.ID), tc.Summary),
		})
	case update.Done:
		conn.SessionUpdate(ctx, acpsdk.SessionNotification{
			SessionId: sessionID,
			Update:    acpsdk.SessionUpdate{},
		})
	case update.Error != "":
		conn.SessionUpdate(ctx, acpsdk.SessionNotification{
			SessionId: sessionID,
			Update:    acpsdk.UpdateAgentMessageText(update.Error),
		})
	}
}

func (b *agentBridge) ResumeSession(ctx context.Context, params acpsdk.ResumeSessionRequest) (acpsdk.ResumeSessionResponse, error) {
	err := b.agent.ResumeSession(ResumeSessionRequest{
		SessionID: SessionID(params.SessionId),
		Cwd:       params.Cwd,
		Meta:      params.Meta,
	})
	return acpsdk.ResumeSessionResponse{}, err
}

func (b *agentBridge) SetSessionConfigOption(ctx context.Context, params acpsdk.SetSessionConfigOptionRequest) (acpsdk.SetSessionConfigOptionResponse, error) {
	return acpsdk.SetSessionConfigOptionResponse{}, nil
}

func (b *agentBridge) SetSessionMode(ctx context.Context, params acpsdk.SetSessionModeRequest) (acpsdk.SetSessionModeResponse, error) {
	return acpsdk.SetSessionModeResponse{}, nil
}
