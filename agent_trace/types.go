package agent_trace

import "encoding/json"

type AgentTraceMetadata struct {
	ID            string   `json:"id"`
	Command       string   `json:"command"`
	CommandArgs   []string `json:"command_args,omitempty"`
	CommandLine   string   `json:"command_line,omitempty"`
	TopicPath     string   `json:"topic_path,omitempty"`
	Workspace     string   `json:"workspace,omitempty"`
	OutputPath    string   `json:"output_path,omitempty"`
	ResumeCommand string   `json:"resume_command,omitempty"`
	ProviderID    string   `json:"provider_id,omitempty"`
	Model         string   `json:"model,omitempty"`
	Status        string   `json:"status"`
	Tags          []string `json:"tags,omitempty"`
	Error         string   `json:"error,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	PromptPath    string   `json:"prompt_path"`
	LogPath       string   `json:"log_path"`
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

type AgentTraceMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content"`
	Sources    []string            `json:"sources,omitempty"`
	ToolCall   *AgentTraceActivity `json:"tool_call,omitempty"`
	StartedAt  *int64              `json:"started_at,omitempty"`
	FinishedAt *int64              `json:"finished_at,omitempty"`
}

type AgentTraceActivity struct {
	Subtype        string                 `json:"subtype"`
	CallID         string                 `json:"call_id,omitempty"`
	ToolName       string                 `json:"tool_name"`
	Summary        string                 `json:"summary"`
	Kind           string                 `json:"kind,omitempty"`
	Status         string                 `json:"status,omitempty"`
	FileChanges    []AgentTraceFileChange `json:"file_changes,omitempty"`
	ReplaceSummary bool                   `json:"replace_summary,omitempty"`
}

type AgentTraceFileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

type Message = AgentTraceMessage
type ToolCallEvent = AgentTraceActivity
type FileChange = AgentTraceFileChange
