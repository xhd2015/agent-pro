package codex_types

import (
	"encoding/json"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type EventType string

const (
	EventStarted   EventType = "item.started"
	EventUpdated   EventType = "item.updated"
	EventCompleted EventType = "item.completed"
	EventError     EventType = "error"
)

type ItemType string

const (
	ItemReasoning        ItemType = "reasoning"
	ItemCommandExecution ItemType = "command_execution"
	ItemFileChange       ItemType = "file_change"
	ItemMessage          ItemType = "message"
)

type Event struct {
	Type    EventType          `json:"type"`
	Item    *EventItem         `json:"item,omitempty"`
	Message string             `json:"message,omitempty"`
	Text    string             `json:"text,omitempty"`
	Mock    *types.MockConfig  `json:"mock,omitempty"`
}

type EventItem struct {
	ID               string          `json:"id"`
	Type             ItemType        `json:"type"`
	Text             string          `json:"text,omitempty"`
	Message          string          `json:"message,omitempty"`
	Content          []ItemPart      `json:"content,omitempty"`
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

type ItemPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type FileChange = types.FileChange

type CodexEvent struct {
	Type    string     `json:"type"`
	Item    *EventItem `json:"item,omitempty"`
	Delta   string     `json:"delta,omitempty"`
	Text    string     `json:"text,omitempty"`
	Message string     `json:"message,omitempty"`
}
