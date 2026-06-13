# Agent Hub Implementation Requirements

## Current State

The agent-hub CLI at `cmd/agent-hub/main.go` already implements:
- `notify` (--json, --file) — produce events
- `hook notify` (--runner, --event) — normalize runner-specific hooks
- `fetch` (--consumer-id, --limit, --peek) — consume events with cursor
- `replay` (--consumer-id, --from partition:offset) — reset consumer cursor
- `sessions` — list all sessions
- `daemon` — start/stop/status
- `consumers` — list consumers
- `partitions` — list partitions
- `status` — print home

Storage is at `agents/agent-hub/storage/storage.go` — file-based JSONL with date partitioning.

Model is at `agents/agent-hub/model/model.go` — NormalizedEvent, Envelope, Cursor, FetchResponse.

Helper binaries:
- `cmd/fake-opencode/main.go` — deterministic fake opencode runner that can invoke hooks

## What Needs to Be Implemented

### 1. `session show` Subcommand

```
agent-hub session show --runner <runner> --session-id <id>
```

Returns JSON with: `runner`, `runner_session_id`, `status` (running/completed/failed), `last_event` (partition + offset).

The session status is determined from the events history:
- `agent.session.started` → status "running"
- `agent.session.finished` → status "completed"
- `agent.session.failed` → status "failed"

The `runSessions` function in main.go already lists all sessions. You need to add `runSessionShow` that reads session data from the sessions directory under `sessions/<runner>/<session_id>.json`.

Error cases:
- Missing `--runner` → error
- Missing `--session-id` → error
- Session not found → error with "not found"

### 2. `session message send` Subcommand

```
agent-hub session message send --runner <runner> --session-id <id> --text <text>
```

Appends a message to the session's message file. If the session doesn't exist, auto-creates it with status "running". If the session was completed/failed, reactivates to "running".

Returns JSON with: `message` (the message object with id, text, timestamp), `session_status` (should be "running").

Messages are stored as a JSON file in the session directory: `sessions/<runner>/<session_id>/messages.jsonl` (JSONL format, one message per line).

Each message has: `id` (unique, e.g., UUID or timestamp-based), `text`, `created_at`.

Error cases:
- Missing `--runner` → error
- Missing `--session-id` → error
- Missing `--text` → error

### 3. `session message list` Subcommand

```
agent-hub session message list --runner <runner> --session-id <id>
```

Returns: `{"messages": [...]}` with all enqueued messages. Does NOT remove messages (peek-only). The session must already exist.

### 4. `session message pop` Subcommand

```
agent-hub session message pop --runner <runner> --session-id <id>
```

Returns: `{"messages": [...]}` with all enqueued messages, then clears the queue (drains/removes all messages). The session must already exist.

### 5. `agent-hub session` Routing

When `session` is the first argument, it delegates:
- `session show` → `runSessionShow`
- `session message send/list/pop` → corresponding functions

The `agent-hub session` without subcommand or with unknown subcommand should return an error.

## Test Verification Command

```sh
doctest test -v agents/agent-hub/tests
```

All 42 tests must pass.

## Current Test Results (Before Implementation)

- 13 PASS (existing features: notify, hook, fetch, replay, sessions)
- 29 FAIL (missing features: session show, session message, fake-opencode hooks)

## Test Tree Structure (Do Not Modify)

Tests are sealed (staged with git add) under:
```
agents/agent-hub/tests/
├── DOCTEST.md
├── SETUP.md                           # Root: Request, Response, Run, helpers
├── produce-events/                    # 10 leaves
├── queue-messages/                    # 13 leaves (ALL FAILING)
├── consume-events/                    # 10 leaves (4 failing)
└── query-session/                     # 9 leaves (8 failing)
```

Tests must NOT be modified. Only implement the production code in:
- `cmd/agent-hub/main.go`
- `agents/agent-hub/storage/storage.go` (if needed)
- `agents/agent-hub/model/model.go` (if needed)

## Important Notes

- All commands read `AGENT_HUB_HOME` env var for storage root
- Session data is stored under `sessions/<runner>/<session_id>/`
- Messages are JSONL format (one JSON object per line in messages.jsonl)
- `cmd/fake-opencode` already supports `--mock-config` with hooks defined in `model.MockConfig`
- The `notifyEvent` helper in tests uses the notify command to produce events
- The `runAgentHub` helper runs agent-hub with given args and captures output
- Tests use `t.TempDir()` for isolation via `AGENT_HUB_HOME`

## Detailed Current Failure Analysis

All 13 queue-message tests fail with: `agent-hub: unknown command: session`
— This is because `session` is not yet handled in the main switch in `run()`.

All query-session failures except `list-all` fail with the same error.
— `list-all` passes because it uses the existing `sessions` command.

The 4 failing fake-opencode tests fail because:
- `fake-opencode-env-var-default`: event stored with wrong runner or session directory check fails
- `fake-opencode-env-var-redirect`: similar
- `fake-opencode-prompt-submitted`: the prompt.submitted event might not have the right format
- `fake-opencode-session-lifecycle`: session creation from hooks might not work

The 4 failing consume-events tests (fetch-cross-partition, fetch-empty, fetch-peek, replay-bad-format) may be due to test setup issues or edge cases in the existing fetch/peek implementation.
