package traceview

import (
	"encoding/json"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

type AgentTraceMetadata struct {
	ID              string            `json:"id"`
	Command         string            `json:"command"`
	CommandArgs     []string          `json:"command_args,omitempty"`
	CommandLine     string            `json:"command_line,omitempty"`
	TopicPath       string            `json:"topic_path,omitempty"`
	Workspace       string            `json:"workspace,omitempty"`
	OutputPath      string            `json:"output_path,omitempty"`
	ResumeCommand   string            `json:"resume_command,omitempty"`
	AgentRunnerID   string            `json:"agent_runner_id,omitempty"`
	ProviderID      string            `json:"provider_id,omitempty"`
	Model           string            `json:"model,omitempty"`
	ParentTraceID   string            `json:"parent_trace_id,omitempty"`
	ParentTraceDir  string            `json:"parent_trace_dir,omitempty"`
	ParentSessionID string            `json:"parent_session_id,omitempty"`
	DelegationID    string            `json:"delegation_id,omitempty"`
	DelegationLabel string            `json:"delegation_label,omitempty"`
	CodexThreadID   string            `json:"codex_thread_id,omitempty"`
	Status          string            `json:"status"`
	Tags            []string          `json:"tags,omitempty"`
	Error           string            `json:"error,omitempty"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
	PromptPath      string            `json:"prompt_path"`
	LogPath         string            `json:"log_path"`
	Children        []AgentTraceChild `json:"children,omitempty"`
}

type AgentTraceChild struct {
	ID              string `json:"id"`
	Command         string `json:"command"`
	CommandLine     string `json:"command_line,omitempty"`
	Status          string `json:"status"`
	AgentRunnerID   string `json:"agent_runner_id,omitempty"`
	Model           string `json:"model,omitempty"`
	CreatedAt       string `json:"created_at"`
	DelegationID    string `json:"delegation_id,omitempty"`
	DelegationLabel string `json:"delegation_label,omitempty"`
}

type AgentTraceSummary struct {
	AgentTraceMetadata
	LogLineCount int `json:"log_line_count"`
}

type AgentTraceDetail struct {
	Metadata AgentTraceMetadata  `json:"metadata"`
	Prompt   string              `json:"prompt"`
	Messages []AgentTraceMessage `json:"messages"`
	RawLines []json.RawMessage   `json:"raw_lines"`
}

type AgentTraceUpdate struct {
	Metadata     AgentTraceMetadata  `json:"metadata"`
	Messages     []AgentTraceMessage `json:"messages"`
	RawLineCount int                 `json:"raw_line_count"`
}

type MessageRole string

const (
	RoleAssistant MessageRole = "assistant"
	RoleToolCall  MessageRole = "tool_call"
)

type ActivitySubtype string

const (
	SubtypeStarted   ActivitySubtype = "started"
	SubtypeUpdated   ActivitySubtype = "updated"
	SubtypeCompleted ActivitySubtype = "completed"
)

type ActivityStatus string

const (
	StatusCompleted  ActivityStatus = "completed"
	StatusFailed     ActivityStatus = "failed"
	StatusInProgress ActivityStatus = "in_progress"
	StatusWarning    ActivityStatus = "warning"
	StatusPending    ActivityStatus = "pending"
)

type AgentTraceMessage struct {
	Role       MessageRole         `json:"role"`
	Content    string              `json:"content"`
	Sources    []string            `json:"sources,omitempty"`
	ToolCall   *AgentTraceActivity `json:"tool_call,omitempty"`
	StartedAt  *int64              `json:"started_at,omitempty"`
	FinishedAt *int64              `json:"finished_at,omitempty"`
}

type AgentTraceActivity struct {
	Subtype        ActivitySubtype        `json:"subtype"`
	CallID         string                 `json:"call_id,omitempty"`
	ToolName       string                 `json:"tool_name"`
	Summary        string                 `json:"summary"`
	Kind           string                 `json:"kind,omitempty"`
	Status         ActivityStatus         `json:"status,omitempty"`
	FileChanges    []AgentTraceFileChange `json:"file_changes,omitempty"`
	ReplaceSummary bool                   `json:"replace_summary,omitempty"`
}

type AgentTraceFileChange = eventtypes.FileChange
type FileChange = eventtypes.FileChange

type AgentTraceParsedEvent struct {
	Message  *AgentTraceMessage
	Activity *AgentTraceActivity
}