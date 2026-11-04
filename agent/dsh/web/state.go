package web

import (
	"context"
	"strings"
)

// State is a durable DSH conversation binding, not a business conversation.
// JSON names retain compatibility with the original Marcus binding records.
// Runner is a discriminator so a shared storage location cannot turn an existing
// terminal conversation into a DSH session. No launch tokens belong here.
type State struct {
	InitialModelSelection *ModelSelection `json:"initial_model_selection,omitempty"`
	Runner                string          `json:"runner"`
	Browser               string          `json:"browser,omitempty"`
	Workspace             string          `json:"workspace,omitempty"`
	DSHSessionID          string          `json:"dsh_session_id,omitempty"`
	Created               bool            `json:"created,omitempty"`
	InitialDelivered      bool            `json:"initial_delivered,omitempty"`
	BrowserOpened         bool            `json:"browser_opened,omitempty"`
	PendingRequestID      string          `json:"pending_request_id,omitempty"`
	PendingHash           string          `json:"pending_hash,omitempty"`
	PendingPrompt         string          `json:"pending_prompt,omitempty"`
	PendingNoSubmit       bool            `json:"pending_no_submit,omitempty"`
}

// IsRunner accepts the UI alias as well as the canonical transport name.
func IsRunner(runner string) bool {
	runner = strings.ToLower(strings.TrimSpace(runner))
	return runner == Runner || runner == "dsh"
}

// Record permits multiple durable checkpoints while its store holds a session
// lock. Implementations must leave the last committed checkpoint intact on error.
type Record interface {
	Load() (State, bool, error)
	Save(State) error
}

// Store separates application-chosen persistence from DSH orchestration.
// WithSession serializes a key across all instances sharing the same storage;
// Read observes the last atomic checkpoint without taking a delivery lock.
type Store interface {
	Read(context.Context, string) (State, bool, error)
	WithSession(context.Context, string, func(Record) error) error
}

// SessionOptions identifies one caller-owned conversation and its first input.
// Workspace/Browser are pinned on first creation; later preference changes do
// not silently relocate an existing session. NoOpen suppresses automatic browser
// opening, but never changes prompt delivery semantics.
type SessionOptions struct {
	// ModelSelection is pinned only when a new binding is reserved.
	ModelSelection *ModelSelection
	Key            string
	Workspace      string
	Browser        string
	Prompt         string
	NoSubmit       bool
	NoOpen         bool
}

type SendOptions struct {
	Key      string
	Prompt   string
	NoSubmit bool
}

// Result distinguishes input admission from browser opening. A browser failure
// after admission is a warning, not a delivery error. Libraries never print logs.
type Result struct {
	SessionID     string
	Created       bool // a provider session was created during this operation
	Accepted      bool // submitted input or a draft was accepted (not agent completion)
	BrowserOpened bool // this operation opened a browser
	Warnings      []Warning
}

type Warning struct {
	Code    string
	Message string
}
