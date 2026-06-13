# Refactor: Move Universal Event Types to agent/event/

## Goal

Move neutral event types from `pkgs/fake-agent/events/` to `agent/event/types/` so other packages can depend on them without coupling to fake-agent. Also split codex and opencode conversion functions into separate packages. Consolidate the duplicated `FileChange` type into a single definition.

## Package Layout After Move

```
agent/event/
  types/types.go              ← ActionType, AgentEvent, FileChange (single source of truth)
  codex_types/convert.go      ← ToCodex() + 5 helper functions
  opencode_types/convert.go   ← ToOpencode() + 1 helper function
```

## Detailed Changes Required

### 1. Create `agent/event/types/types.go`

Export the neutral types (move from `pkgs/fake-agent/events/events.go`):
- `ActionType` string type
- Constants: `ActionThink`, `ActionToolCall`, `ActionMessage`, `ActionError`, `ActionDone`
- `AgentEvent` struct with all fields and JSON tags
- `FileChange` struct with `Path string` and `Kind string`

This package imports only the standard library. No imports from `pkgs/` or `agent/`.

### 2. Create `agent/event/codex_types/convert.go`

Move the `ToCodex` function and its helpers:
- `ToCodex(events []types.AgentEvent) []fakeagent.Event`
- `convertToolCallToCodex(e types.AgentEvent, id string) []fakeagent.Event`
- `convertBashToCodex(e types.AgentEvent, id string) []fakeagent.Event`
- `convertReadToCodex(e types.AgentEvent, id string) []fakeagent.Event`
- `convertWriteToCodex(e types.AgentEvent, id string) []fakeagent.Event`
- `convertGrepToCodex(e types.AgentEvent, id string) []fakeagent.Event`

Imports: `agent/event/types`, `pkgs/fake-agent`, `pkgs/fake-agent/fake-tool-exec`, `fmt`, `strings`.

### 3. Create `agent/event/opencode_types/convert.go`

Move the `ToOpencode` function and its helper:
- `ToOpencode(events []types.AgentEvent, sessionID string) []map[string]any`
- `convertToolCallToOpencode(e types.AgentEvent, id, sessionID string) map[string]any`

Imports: `agent/event/types`, `pkgs/fake-agent/fake-tool-exec`, `fmt`, `strings`.

### 4. Update `pkgs/fake-agent/events/events.go`

Remove the moved types and functions. The file may become empty or contain only what remains (the generator is in a separate file `generator.go`).

### 5. Consolidate `FileChange` duplication

There are currently three identical `FileChange` definitions:
- `pkgs/fake-agent/events/events.go` (will be removed as part of move)
- `pkgs/fake-agent/fakeagent.go` — replace with import from `agent/event/types`
- `pkgs/fake-agent/fake-tool-exec/fake_tool_exec.go` — replace with import from `agent/event/types`

All consumers should reference `types.FileChange` from `agent/event/types`.

### 6. Update imports in callers

- `cmd/fake-codex/main.go` — update import paths for `events.ToCodex` and `events.AgentEvent`
- `cmd/fake-opencode/main.go` — update import paths for `events.ToOpencode`, `events.AgentEvent`, `events.GenerateEvents`
- `pkgs/fake-agent/events/generator.go` — update import for `AgentEvent`, `ActionType` (import from `agent/event/types`)

### 7. Update `pkgs/fake-agent/fakeagent.go`

Replace local `FileChange` type (line 55-58) with reference to `types.FileChange` from `agent/event/types`. The `EventItem.Changes` field type becomes `[]types.FileChange`. Update all usages to import appropriately.

### 8. Update `pkgs/fake-agent/fake-tool-exec/fake_tool_exec.go`

Replace local `FileChange` type (line 18-21) with reference to `types.FileChange` from `agent/event/types`. The `MockConfig.Changes` field type becomes `[]types.FileChange`.

## Important Constraints

- **No behavior change**: The conversion logic, constants, and struct fields must remain exactly as they are. This is a pure move + dedup refactoring.
- **Tests are sealed**: All test files under `agent/event/tests/` are staged (git added) and must NOT be modified. Any implementation must make these tests pass as-is.
- **Existing tests must still pass**: All existing doctests under `cmd/fake-codex/tests/` and `cmd/fake-opencode/tests/` must continue to pass.
- **`generator.go` stays**: `pkgs/fake-agent/events/generator.go` stays where it is (only import paths update).

## Sealed Test Tree

```
agent/event/tests/
├── DOCTEST.md
├── SETUP.md                              ← builds+run a Go program importing new packages
├── types-action-constants/               ← verifies ActionType constants exist
├── types-agent-event-fields/             ← verifies AgentEvent JSON marshaling
├── types-file-change-fields/             ← verifies FileChange JSON marshaling
├── codex_types-think/                    ← ToCodex: think → codex reasoning events
├── codex_types-message/                  ← ToCodex: message → codex message event
├── codex_types-error/                    ← ToCodex: error → codex error event
├── codex_types-tool-bash-mock/           ← ToCodex: tool_call+bash+mock → codex events
├── opencode_types-think/                 ← ToOpencode: think → opencode reasoning event
├── opencode_types-message/               ← ToOpencode: message → opencode text event
├── opencode_types-error/                 ← ToOpencode: error → opencode error event
├── opencode_types-done/                  ← ToOpencode: done → opencode done event
└── opencode_types-tool-bash-mock/        ← ToOpencode: tool_call+bash+mock → opencode event
```

12 leaf tests. Each test writes a small `main.go` that imports from the new packages, calls the relevant function, and prints JSON output. ASSERT.md checks the output.

## Verification

Run to confirm all tests pass:
```
doctest test -v ./agent/event/tests
```

Also verify existing tests still pass:
```
doctest test -v ./cmd/fake-codex/tests ./cmd/fake-opencode/tests
```
