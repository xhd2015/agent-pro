package acp

type SessionID string

type AgentInfo struct {
	Name    string
	Version string
}

type SessionInfo struct {
	SessionID SessionID
	Cwd       string
	Title     string
	UpdatedAt string
}

type NewSessionRequest struct {
	Cwd  string
	Meta map[string]any
}

type ResumeSessionRequest struct {
	SessionID SessionID
	Cwd       string
	Meta      map[string]any
}

type ListSessionsRequest struct {
	Cwd    string
	Cursor string
}

type ListSessionsResponse struct {
	Sessions   []SessionInfo
	NextCursor string
}

type PromptRequest struct {
	SessionID SessionID
	Content   string
	Meta      map[string]any
}

type AuthenticateRequest struct {
	Token string
}

type CancelSessionRequest struct {
	SessionID SessionID
}

type CloseSessionRequest struct {
	SessionID SessionID
}

type SessionUpdate struct {
	Text      string
	Done      bool
	ToolCall  *ToolCallUpdate
	Error     string
}

type ToolCallUpdate struct {
	ID      string
	Tool    string
	Status  string
	Summary string
	Output  string
	Error   string
}

type Agent interface {
	Info() AgentInfo
	Authenticate(ctx AuthenticateRequest) error
	NewSession(req NewSessionRequest) (SessionID, error)
	ResumeSession(req ResumeSessionRequest) error
	ListSessions(req ListSessionsRequest) (ListSessionsResponse, error)
	CloseSession(req CloseSessionRequest) error
	CancelSession(req CancelSessionRequest) error
	Prompt(req PromptRequest, onUpdate func(SessionUpdate)) error
}
