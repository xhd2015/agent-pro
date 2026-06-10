package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const SchemaVersionEventV1 = "agent-hub.event.v1"

type EventType string

const (
	EventSessionStarted    EventType = "agent.session.started"
	EventPromptSubmitted   EventType = "agent.prompt.submitted"
	EventSessionUpdated    EventType = "agent.session.updated"
	EventSessionFinished   EventType = "agent.session.finished"
	EventSessionFailed     EventType = "agent.session.failed"
	EventToolStarted       EventType = "agent.tool.started"
	EventToolFinished      EventType = "agent.tool.finished"
	EventPermissionAsked   EventType = "agent.permission.requested"
	EventPermissionReplied EventType = "agent.permission.replied"
)

var validEventTypes = map[EventType]bool{
	EventSessionStarted:    true,
	EventPromptSubmitted:   true,
	EventSessionUpdated:    true,
	EventSessionFinished:   true,
	EventSessionFailed:     true,
	EventToolStarted:       true,
	EventToolFinished:      true,
	EventPermissionAsked:   true,
	EventPermissionReplied: true,
}

type NormalizedEvent struct {
	EventType       EventType       `json:"event_type"`
	Runner          string          `json:"runner"`
	RunnerSessionID string          `json:"runner_session_id,omitempty"`
	Workspace       string          `json:"workspace,omitempty"`
	Model           string          `json:"model,omitempty"`
	Prompt          string          `json:"prompt,omitempty"`
	OccurredAt      time.Time       `json:"occurred_at,omitempty"`
	Payload         json.RawMessage `json:"payload,omitempty"`
}

func (e NormalizedEvent) Validate() error {
	if strings.TrimSpace(string(e.EventType)) == "" {
		return fmt.Errorf("event_type is required")
	}
	if !validEventTypes[e.EventType] {
		return fmt.Errorf("unknown event_type: %s", e.EventType)
	}
	if strings.TrimSpace(e.Runner) == "" {
		return fmt.Errorf("runner is required")
	}
	return nil
}

type Producer struct {
	CLIVersion string `json:"cli_version,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	PID        int    `json:"pid,omitempty"`
}

type Envelope struct {
	SchemaVersion string          `json:"schema_version"`
	EventID       string          `json:"event_id"`
	Partition     string          `json:"partition"`
	Offset        int64           `json:"offset"`
	ReceivedAt    time.Time       `json:"received_at"`
	Producer      Producer        `json:"producer,omitempty"`
	Event         NormalizedEvent `json:"event"`
}

func (e Envelope) Validate() error {
	if strings.TrimSpace(e.SchemaVersion) == "" {
		return fmt.Errorf("schema_version is required")
	}
	if e.SchemaVersion != SchemaVersionEventV1 {
		return fmt.Errorf("unknown schema_version: %s", e.SchemaVersion)
	}
	if strings.TrimSpace(e.EventID) == "" {
		return fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(e.Partition) == "" {
		return fmt.Errorf("partition is required")
	}
	if err := validatePartition(e.Partition); err != nil {
		return err
	}
	if e.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	if e.ReceivedAt.IsZero() {
		return fmt.Errorf("received_at is required")
	}
	if err := e.Event.Validate(); err != nil {
		return fmt.Errorf("event: %w", err)
	}
	return nil
}

type Cursor struct {
	Partition string `json:"partition"`
	Offset    int64  `json:"offset"`
}

func (c Cursor) Validate() error {
	if strings.TrimSpace(c.Partition) == "" {
		return fmt.Errorf("partition is required")
	}
	if err := validatePartition(c.Partition); err != nil {
		return err
	}
	if c.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	return nil
}

type FetchResponse struct {
	ConsumerID     string     `json:"consumer_id"`
	Events         []Envelope `json:"events"`
	PreviousCursor Cursor     `json:"previous_cursor"`
	NextCursor     Cursor     `json:"next_cursor"`
	HasMore        bool       `json:"has_more"`
}

type MockConfig struct {
	Version          string          `json:"version"`
	Runner           string          `json:"runner"`
	SessionID        string          `json:"session_id,omitempty"`
	Model            string          `json:"model,omitempty"`
	DelayMS          int             `json:"delay_ms,omitempty"`
	ExitCode         int             `json:"exit_code,omitempty"`
	Stderr           string          `json:"stderr,omitempty"`
	IgnoreHookErrors bool            `json:"ignore_hook_errors,omitempty"`
	HookCommand      string          `json:"hook_command,omitempty"`
	StdoutEvents     json.RawMessage `json:"stdout_events,omitempty"`
	Hooks            []MockHook      `json:"hooks,omitempty"`
}

type MockHook struct {
	At      string          `json:"at"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

var partitionPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func validatePartition(partition string) error {
	if !partitionPattern.MatchString(partition) {
		return fmt.Errorf("partition must use YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", partition); err != nil {
		return fmt.Errorf("partition must use YYYY-MM-DD")
	}
	return nil
}
