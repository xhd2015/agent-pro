package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

type AgentTraceFileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

type Message = AgentTraceMessage
type ToolCallEvent = AgentTraceActivity
type FileChange = AgentTraceFileChange

type AgentTraceParsedEvent struct {
	Message  *AgentTraceMessage
	Activity *AgentTraceActivity
}

type AgentTraceAdapter interface {
	Name() string
	Parse(raw json.RawMessage) (AgentTraceParsedEvent, bool)
}

type registeredAgentTraceAdapter struct {
	priority int
	adapter  AgentTraceAdapter
}

var agentTraceAdapters []registeredAgentTraceAdapter

func RegisterAgentTraceAdapter(priority int, adapter AgentTraceAdapter) {
	if adapter == nil {
		return
	}
	agentTraceAdapters = append(agentTraceAdapters, registeredAgentTraceAdapter{
		priority: priority,
		adapter:  adapter,
	})
	sort.SliceStable(agentTraceAdapters, func(i, j int) bool {
		return agentTraceAdapters[i].priority < agentTraceAdapters[j].priority
	})
}

func ParseAgentTraceLine(raw json.RawMessage) (AgentTraceParsedEvent, bool) {
	for _, registered := range agentTraceAdapters {
		if parsed, ok := registered.adapter.Parse(raw); ok {
			return parsed, true
		}
	}
	return AgentTraceParsedEvent{}, false
}

type TraceEvent struct {
	Type     string                     `json:"type"`
	Subtype  string                     `json:"subtype,omitempty"`
	CallID   string                     `json:"call_id,omitempty"`
	Message  *TraceMessage              `json:"message,omitempty"`
	Result   string                     `json:"result,omitempty"`
	Item     *TraceItem                 `json:"item,omitempty"`
	ToolCall map[string]json.RawMessage `json:"tool_call,omitempty"`
	Delta    string                     `json:"delta,omitempty"`
	Text     string                     `json:"text,omitempty"`
}

type TraceMessage struct {
	Content []TraceContent `json:"content"`
}

type TraceContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

type TraceItem struct {
	ID               string          `json:"id,omitempty"`
	Type             string          `json:"type,omitempty"`
	Text             string          `json:"text,omitempty"`
	Message          string          `json:"message,omitempty"`
	Content          []TraceContent  `json:"content,omitempty"`
	Command          string          `json:"command,omitempty"`
	AggregatedOutput string          `json:"aggregated_output,omitempty"`
	ExitCode         *int            `json:"exit_code,omitempty"`
	Status           string          `json:"status,omitempty"`
	Items            []TraceTodoItem `json:"items,omitempty"`
	Plan             []TracePlanItem `json:"plan,omitempty"`
	Explanation      string          `json:"explanation,omitempty"`
	Changes          []FileChange    `json:"changes,omitempty"`
	Raw              json.RawMessage `json:"-"`
}

type TraceTodoItem struct {
	Text      string `json:"text,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Status    string `json:"status,omitempty"`
}

type TracePlanItem struct {
	Step   string `json:"step,omitempty"`
	Status string `json:"status,omitempty"`
}

func (i *TraceItem) UnmarshalJSON(data []byte) error {
	type alias TraceItem
	var parsed alias
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*i = TraceItem(parsed)
	i.Raw = append(i.Raw[:0], data...)
	return nil
}

func TraceIsAssistantItem(itemType string) bool {
	switch itemType {
	case "", "agent_message", "message", "assistant_message", "output_text":
		return true
	default:
		return false
	}
}

func TraceItemText(item *TraceItem) string {
	if item == nil {
		return ""
	}
	if strings.TrimSpace(item.Text) != "" {
		return item.Text
	}
	if strings.TrimSpace(item.Message) != "" {
		return item.Message
	}
	var b strings.Builder
	for _, part := range item.Content {
		if part.Text != "" {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func TraceMessageText(message *TraceMessage) string {
	if message == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range message.Content {
		if part.Type == "" || part.Type == "text" {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

func CompactTraceOutput(output string) string {
	if len([]byte(output)) <= 4000 && strings.Count(output, "\n") <= 40 {
		return output
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	head := traceJoinLines(lines, 0, minTraceInt(10, len(lines)))
	tailStart := len(lines) - minTraceInt(10, len(lines))
	if tailStart < 10 {
		tailStart = 10
	}
	tail := traceJoinLines(lines, tailStart, len(lines))
	return fmt.Sprintf("[omitted: %d lines, %d bytes]\n--- first lines ---\n%s\n--- last lines ---\n%s", len(lines), len([]byte(output)), head, tail)
}

func traceJoinLines(lines []string, start, end int) string {
	if start >= end {
		return ""
	}
	out := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		out = append(out, truncateTraceLine(line))
	}
	return strings.Join(out, "\n")
}

func truncateTraceLine(line string) string {
	runes := []rune(line)
	if len(runes) <= 260 {
		return line
	}
	return string(runes[:260]) + "...<line truncated>"
}

func minTraceInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TitleFromIdentifier(id string) string {
	if strings.TrimSpace(id) == "" {
		return "Tool"
	}
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
