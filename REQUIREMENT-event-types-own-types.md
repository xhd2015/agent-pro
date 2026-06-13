# Refactor: Own Types in codex_types and opencode_types

## Goal

Two improvements to the `agent/event/` packages:

1. **`codex_types` should own its types** — `convert.go` currently imports `fakeagent.Event`, `fakeagent.EventItem`, `fakeagent.EventType`, `fakeagent.ItemType`. These types should be defined directly in `codex_types` so the package is self-contained.

2. **`opencode_types` should use typed structs** — `convert.go` currently returns `[]map[string]any`. This should be replaced with proper typed structs.

## Detailed Changes

### Part 1: codex_types — own its types

Move the following types from `pkgs/fake-agent/fakeagent.go` into a new file `agent/event/codex_types/types.go`:

```go
package codex_types

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
    Type    EventType      `json:"type"`
    Item    *EventItem     `json:"item,omitempty"`
    Message string         `json:"message,omitempty"`
    Text    string         `json:"text,omitempty"`
    Mock    *MockConfig    `json:"mock,omitempty"`
}

type EventItem struct {
    ID               string       `json:"id"`
    Type             ItemType     `json:"type"`
    Text             string       `json:"text,omitempty"`
    Content          []ItemPart   `json:"content,omitempty"`
    Command          string       `json:"command,omitempty"`
    AggregatedOutput string       `json:"aggregated_output,omitempty"`
    ExitCode         *int         `json:"exit_code,omitempty"`
    Status           string       `json:"status,omitempty"`
    Changes          []FileChange `json:"changes,omitempty"`
}

type ItemPart struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

type FileChange = types.FileChange  // re-export from agent/event/types

type CodexEvent struct {
    Type    string     `json:"type"`
    Item    *EventItem `json:"item,omitempty"`
    Delta   string     `json:"delta,omitempty"`
    Text    string     `json:"text,omitempty"`
    Message string     `json:"message,omitempty"`
}
```

The `Mock` field on `Event` currently references `faketoolexec.MockConfig` in fakeagent. Since `faketoolexec.MockConfig` is already a type alias to `types.MockConfig`, use `types.MockConfig` directly.

**Update `agent/event/codex_types/convert.go`:**
- Remove import of `fakeagent`
- Use local types (`Event`, `EventItem`, `EventStarted`, etc.) instead of `fakeagent.Event`, `fakeagent.EventStarted`, etc.
- `ToCodex` returns `[]Event` (local, not `[]fakeagent.Event`)
- `FileChange` references: use `types.FileChange` or `FileChange` from local types.go

**Update `pkgs/fake-agent/fakeagent.go`:**
- Replace `Event`, `EventItem`, `EventType`, `ItemType`, `ItemPart` definitions with type aliases to `codex_types`:
  ```go
  type EventType = codex_types.EventType
  type Event = codex_types.Event
  type EventItem = codex_types.EventItem
  type ItemType = codex_types.ItemType
  type ItemPart = codex_types.ItemPart
  ```
- Keep all consts as aliases too, or forward-declare them:
  ```go
  const (
      EventStarted   = codex_types.EventStarted
      EventUpdated   = codex_types.EventUpdated
      EventCompleted = codex_types.EventCompleted
      EventError     = codex_types.EventError
  )
  const (
      ItemReasoning        = codex_types.ItemReasoning
      ItemCommandExecution = codex_types.ItemCommandExecution
      ItemFileChange       = codex_types.ItemFileChange
      ItemMessage          = codex_types.ItemMessage
  )
  ```

**Move `CodexEvent` from `pkgs/fake-agent/codex.go` to `agent/event/codex_types/types.go`**

### Part 2: opencode_types — typed structs

Create `agent/event/opencode_types/types.go`:

```go
package opencode_types

type Event struct {
    Type      string       `json:"type"`
    SessionID string       `json:"sessionID,omitempty"`
    Part      any          `json:"part,omitempty"`
    Error     *ErrorDetail `json:"error,omitempty"`
    Done      bool         `json:"done,omitempty"`
}

type ReasoningPart struct {
    ID   string `json:"id"`
    Type string `json:"type"`
    Text string `json:"text"`
}

type TextPart struct {
    ID   string `json:"id"`
    Type string `json:"type"`
    Text string `json:"text"`
}

type ErrorDetail struct {
    Name string     `json:"name"`
    Data *ErrorData `json:"data"`
}

type ErrorData struct {
    Message string `json:"message"`
}

type ToolUsePart struct {
    ID    string       `json:"id"`
    Type  string       `json:"type"`
    Tool  string       `json:"tool"`
    State ToolUseState `json:"state"`
}

type ToolUseState struct {
    Input    map[string]any `json:"input,omitempty"`
    Output   string         `json:"output,omitempty"`
    Stderr   string         `json:"stderr,omitempty"`
    ExitCode int            `json:"exit_code,omitempty"`
    Error    string         `json:"error,omitempty"`
    Status   string         `json:"status,omitempty"`
}
```

**Update `agent/event/opencode_types/convert.go`:**
- `ToOpencode` returns `[]Event` (typed) instead of `[]map[string]any`
- `convertToolCallToOpencode` returns `Event` instead of `map[string]any`
- Replace all `map[string]any` literals with typed struct constructors
- JSON marshaling behavior must be identical (same field names, same omitempty, same nesting)

**Update `cmd/fake-opencode/main.go`:**
- Replace `map[string]any` event handling with typed struct field access
- `stdoutEvents []map[string]any` → `stdoutEvents []opencode_types.Event`
- `withSession`, `resolveOpencodeStdoutEvents`, `processToolUseEvent`, `applyMockOutput`, `parseMockFromMap` — all updated to use typed structs
- Any other `map[string]any` used for opencode events should use typed struct access

## Important Constraints

- **No behavior change**: JSON output must be byte-for-byte identical. `omitempty` must match exactly what map encoders produce.
- **Tests are sealed**: All 16 test files under `agent/event/tests/` are sealed (staged or committed) and must NOT be modified.
- **Existing tests must still pass**: All doctests under `cmd/fake-codex/tests/` and `cmd/fake-opencode/tests/` must continue to pass.
- **`fakeagent` backward compat**: `fakeagent.Event`, `fakeagent.GenerateSession()`, `fakeagent.FormatCodexEvents()`, etc. must continue to work with the same types — achieved via type aliases.
- **No circular imports**: `codex_types` imports `types`; `fakeagent` imports `codex_types`; `opencode_types` imports `types` (and `faketoolexec` for tool execution).

## Sealed Test Tree (16 leaves)

Existing (12):
```
types-action-constants, types-agent-event-fields, types-file-change-fields,
codex_types-think, codex_types-message, codex_types-error, codex_types-tool-bash-mock,
opencode_types-think, opencode_types-message, opencode_types-error, opencode_types-done,
opencode_types-tool-bash-mock
```

New (4) — these are the RED tests the implementation must turn GREEN:
```
codex_types-event-types          ← Event, EventItem, all 8 constants
opencode_types-event-reasoning   ← Event + ReasoningPart JSON output
opencode_types-event-error       ← Event + ErrorDetail JSON output
opencode_types-event-tool        ← Event + ToolUsePart/ToolUseState JSON output
```

## Verification

```
doctest test -v ./agent/event/tests
doctest test -v ./cmd/fake-codex/tests ./cmd/fake-opencode/tests
```
