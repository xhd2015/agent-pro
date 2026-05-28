package registry

import (
	"context"
	"io"
)

type DeltaCallback func(delta string)

type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

type ToolCallEvent struct {
	Subtype        string       `json:"subtype"`
	CallID         string       `json:"call_id,omitempty"`
	ToolName       string       `json:"tool_name"`
	Summary        string       `json:"summary"`
	Kind           string       `json:"kind,omitempty"`
	Status         string       `json:"status,omitempty"`
	FileChanges    []FileChange `json:"file_changes,omitempty"`
	ReplaceSummary bool         `json:"replace_summary,omitempty"`
}

type ToolCallCallback func(event ToolCallEvent)

type AskOptions struct {
	Model            string
	RawLog           io.Writer
	Workspace        string
	AgentMode        bool
	DisableSubAgents bool
	SandboxMode      string
	OnToolCall       ToolCallCallback
}

type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Agent interface {
	Ask(ctx context.Context, question string, opts *AskOptions, onDelta DeltaCallback) (fullAnswer string, err error)
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
