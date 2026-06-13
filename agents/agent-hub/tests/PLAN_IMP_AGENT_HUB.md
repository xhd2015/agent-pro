# Agent Hub — Implementation Test Plan

This document describes the full doctest tree for the agent-hub CLI. All tests
reside under `agents/agent-hub/tests/`. The tree is organized into 4 groups:
**Produce Events**, **Queue Messages**, **Consume Events**, **Query Session**.

---

## Environment

| Variable | Value | Purpose |
|----------|-------|---------|
| `AGENT_HUB_HOME` | `<t.TempDir()>` | Isolate all test storage |
| `AGENT_HUB_OPENCODE_RUNNER` | `fake-opencode` (most tests) or unset (env-var-default test) | Redirect `--runner opencode` to a mock runner name |

---

## Root SETUP.md Design

```go
type Request struct {
    Home          string   // AGENT_HUB_HOME path
    Command       string   // e.g., "notify", "hook", "fetch", "session"
    Args          []string // subcommand + flags, e.g., ["show","--runner","codex","--session-id","s1"]
    Stdin         string   // stdin payload (for hook notify)
    MockConfig    string   // path to mock config file (for fake-opencode tests)
}

type Response struct {
    Stdout   string
    Stderr   string
    ExitCode int
}
```

Root `Run` shells out to `agent-hub` or `fake-opencode`:

```go
func Run(t *testing.T, req *Request) (*Response, error) {
    var cmd *exec.Cmd
    cmd = exec.Command(req.Command, req.Args...)
    cmd.Env = append(os.Environ(), "AGENT_HUB_HOME="+req.Home)
    if req.Stdin != "" {
        cmd.Stdin = strings.NewReader(req.Stdin)
    }
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    resp := &Response{
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        ExitCode: exitCode(err),
    }
    return resp, nil
}
```

Leaves override `func Setup(t *testing.T, req *Request) error` to populate
`req.Command`, `req.Args`, `req.Stdin`, `req.Home` (via `t.TempDir()`), and
if needed create mock config files. Leaves override `func Run` only when the
command under test is `fake-opencode` (not `agent-hub`).

---

## Mock Config Reference

For tests that exercise `fake-opencode run --mock-config`, the config file
uses `llm_events` (unified `AgentEvent` format) and `hooks` (opencode native
event names):

```json
{
  "version": "agent-pro.fake-runner.v1",
  "runner": "fake-opencode",
  "session_id": "sess_001",
  "model": "openai/gpt-5",
  "hook_command": "agent-hub hook notify --runner opencode --event {{event}}",
  "llm_events": [{"type":"done"}],
  "hooks": [
    {"at":"before_stdout","event":"session.created"},
    {"at":"before_exit","event":"session.idle"}
  ]
}
```

- `llm_events` entries are `agent/event/types.AgentEvent` objects: `{"type":"think|tool_call|message|error|done|step_start|step_finish",...}`
- `hooks` entries use opencode native event names: `session.created`, `message.updated`, `session.idle`, `session.error`, etc.
- `hook_command` template uses `{{event}}` placeholder; `fake-opencode` replaces it with the hook's event name and pipes the hook payload to stdin.

---

## Test Tree (42 leaves)

```
tests/
├── DOCTEST.md
├── SETUP.md                           # Root: Request, Response, Run
├── produce-events/
│   ├── SETUP.md                       # Group: produce events via agent-hub
│   ├── notify-json-valid/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── notify-file/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── notify-invalid-event-type/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── notify-missing-input/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── hook-unknown-runner/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── hook-unknown-event/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fake-opencode-session-lifecycle/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fake-opencode-prompt-submitted/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fake-opencode-env-var-redirect/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   └── fake-opencode-env-var-default/
│       ├── SETUP.md
│       └── ASSERT.md
├── queue-messages/
│   ├── SETUP.md                       # Group: session message send/list/pop
│   ├── send-to-running/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-to-completed/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-to-failed/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-to-non-existent/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-missing-runner/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-missing-session-id/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── send-missing-text/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── list-with-messages/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── list-empty/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── pop-with-messages/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── pop-empty/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── roundtrip-single/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   └── roundtrip-multiple/
│       ├── SETUP.md
│       └── ASSERT.md
├── consume-events/
│   ├── SETUP.md                       # Group: consume events via fetch/replay
│   ├── fetch-from-start/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-with-limit/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-less-than-limit/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-peek/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-advances-cursor/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-cross-partition/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-empty/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── fetch-missing-consumer-id/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   ├── replay-reset-cursor/
│   │   ├── SETUP.md
│   │   └── ASSERT.md
│   └── replay-bad-format/
│       ├── SETUP.md
│       └── ASSERT.md
└── query-session/
    ├── SETUP.md                       # Group: session show + sessions listing
    ├── show-running/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-completed/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-failed/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-non-existent/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-missing-runner/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-missing-session-id/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── show-after-transition/
    │   ├── SETUP.md
    │   └── ASSERT.md
    ├── list-all/
    │   ├── SETUP.md
    │   └── ASSERT.md
    └── list-empty/
        ├── SETUP.md
        └── ASSERT.md
```

---

## Leaf Test Cases — Detailed

### Group 1: Produce Events (10 leaves)

#### 1. `produce-events/notify-json-valid/`

- **Setup**: Sets `AGENT_HUB_HOME`, runs `agent-hub notify --json '{"event_type":"agent.session.started","runner":"fake-opencode","runner_session_id":"s1"}'`
- **Assert**: ExitCode 0, stdout JSON has `event_id` (non-empty), `partition` (YYYY-MM-DD), `offset` (≥0), `received_at` (non-empty). Event persisted: run `agent-hub fetch --consumer-id test --limit 1` and verify first event has `runner:"fake-opencode"`.

#### 2. `produce-events/notify-file/`

- **Setup**: Writes a valid NormalizedEvent JSON to temp file, runs `agent-hub notify --file <path>`
- **Assert**: Same as notify-json-valid.

#### 3. `produce-events/notify-invalid-event-type/`

- **Setup**: `agent-hub notify --json '{"event_type":"bogus","runner":"x"}'`
- **Assert**: ExitCode ≠ 0, stderr contains error message.

#### 4. `produce-events/notify-missing-input/`

- **Setup**: `agent-hub notify` (no flags)
- **Assert**: ExitCode ≠ 0, stderr contains `--json` or `--file`.

#### 5. `produce-events/hook-unknown-runner/`

- **Setup**: `agent-hub hook notify --runner unknown --event SessionStart` with stdin payload
- **Assert**: ExitCode ≠ 0, stderr contains `unknown runner`.

#### 6. `produce-events/hook-unknown-event/`

- **Setup**: `agent-hub hook notify --runner opencode --event BogusEvent` with stdin payload
- **Assert**: ExitCode ≠ 0, stderr contains `unknown hook event`.

#### 7. `produce-events/fake-opencode-session-lifecycle/`

- **Setup**: Creates mock config file with `runner:"fake-opencode"`, `session_id:"sess_life"`, hooks `session.created` (before_stdout) + `session.idle` (before_exit). Runs `fake-opencode run --mock-config <path>` with `AGENT_HUB_OPENCODE_RUNNER=fake-opencode`.
- **Assert**: `fake-opencode` exits 0. `agent-hub fetch --consumer-id test --limit 10` returns 2 events: `agent.session.started` + `agent.session.finished`, both with `runner:"fake-opencode"`, `runner_session_id:"sess_life"`. `agent-hub session show --runner fake-opencode --session-id sess_life` returns `status:"completed"`.

#### 8. `produce-events/fake-opencode-prompt-submitted/`

- **Setup**: Mock config with `runner:"fake-opencode"`, `session_id:"sess_prompt"`, hook `message.updated` (before_stdout) with payload `{"message":{"text":"hello world"}}`. Runs `fake-opencode run --mock-config`.
- **Assert**: `agent-hub fetch --consumer-id test --limit 10` includes an `agent.prompt.submitted` event with `prompt:"hello world"`.

#### 9. `produce-events/fake-opencode-env-var-redirect/`

- **Setup**: `AGENT_HUB_OPENCODE_RUNNER=fake-opencode`. Mock config with `runner:"opencode"`, `hook_command:"agent-hub hook notify --runner opencode --event {{event}}"`, hook `session.created`.
- **Assert**: Event stored with `runner:"fake-opencode"` (not `"opencode"`). Session created under `sessions/fake-opencode/`.

#### 10. `produce-events/fake-opencode-env-var-default/`

- **Setup**: `AGENT_HUB_OPENCODE_RUNNER` NOT set. Mock config with `runner:"opencode"`, `hook_command:"agent-hub hook notify --runner opencode --event {{event}}"`.
- **Assert**: Event stored with `runner:"opencode"`. Session under `sessions/opencode/`.

---

### Group 2: Queue Messages (13 leaves)

All use `AGENT_HUB_OPENCODE_RUNNER=fake-opencode`. Session preconditions established via `agent-hub notify` events before testing message commands.

#### 11. `queue-messages/send-to-running/`

- **Setup**: Create running session via `agent.session.started` notify. Run `agent-hub session message send --runner fake-opencode --session-id s1 --text "followup"`.
- **Assert**: Responds with `{"message":{...},"session_status":"running"}`. `agent-hub session show --runner fake-opencode --session-id s1` still returns `status:"running"`.

#### 12. `queue-messages/send-to-completed/`

- **Setup**: Create and complete session (started + finished events). Run `send`.
- **Assert**: `session_status:"running"` in response. `session show` now returns `status:"running"`.

#### 13. `queue-messages/send-to-failed/`

- **Setup**: Create and fail session (started + failed events). Run `send`.
- **Assert**: `session_status:"running"`. `session show` returns `status:"running"`.

#### 14. `queue-messages/send-to-non-existent/`

- **Setup**: No prior session exists. Run `agent-hub session message send --runner fake-opencode --session-id newone --text "hello"`.
- **Assert**: `session_status:"running"`. `session show` returns `status:"running"` with `runner_session_id:"newone"`.

#### 15. `queue-messages/send-missing-runner/`

- **Setup**: `agent-hub session message send --session-id s1 --text "hi"`
- **Assert**: ExitCode ≠ 0, error about `--runner`.

#### 16. `queue-messages/send-missing-session-id/`

- **Setup**: `agent-hub session message send --runner fake-opencode --text "hi"`
- **Assert**: ExitCode ≠ 0, error about `--session-id`.

#### 17. `queue-messages/send-missing-text/`

- **Setup**: `agent-hub session message send --runner fake-opencode --session-id s1`
- **Assert**: ExitCode ≠ 0, error about `--text`.

#### 18. `queue-messages/list-with-messages/`

- **Setup**: Create running session, send 2 messages ("msg1", "msg2"). Run `list`.
- **Assert**: Returns `{"messages":[...]}` with 2 items, `text` fields are `"msg1"` and `"msg2"`. Run `list` again → same 2 items (file intact, peek).

#### 19. `queue-messages/list-empty/`

- **Setup**: Create running session, no messages. Run `list`.
- **Assert**: `{"messages":[]}`.

#### 20. `queue-messages/pop-with-messages/`

- **Setup**: Create running session, send 2 messages. Run `pop`.
- **Assert**: Returns `{"messages":[...]}` with 2 items. Run `pop` again → `{"messages":[]}` (drained).

#### 21. `queue-messages/pop-empty/`

- **Setup**: Create running session, no messages. Run `pop`.
- **Assert**: `{"messages":[]}`.

#### 22. `queue-messages/roundtrip-single/`

- **Setup**: Create running session, `send --text "exact message"`. Run `pop`.
- **Assert**: Single message returned, `text:"exact message"`, `runner:"fake-opencode"`, `session_id` matches.

#### 23. `queue-messages/roundtrip-multiple/`

- **Setup**: Create running session, send 3 messages: "A", "B", "C". Run `pop`.
- **Assert**: 3 messages returned in insertion order (A, B, C).

---

### Group 3: Consume Events (10 leaves)

Events produced via `agent-hub notify` (or fake-opencode for cross-partition).
All use `AGENT_HUB_OPENCODE_RUNNER=fake-opencode`.

#### 24. `consume-events/fetch-from-start/`

- **Setup**: Produce 1 event via notify. Run `agent-hub fetch --consumer-id c1 --limit 10`.
- **Assert**: Returns 1 event. `previous_cursor` is zero. `next_cursor` advanced.

#### 25. `consume-events/fetch-with-limit/`

- **Setup**: Produce 5 events. Run `fetch --consumer-id c1 --limit 3`.
- **Assert**: Returns exactly 3 events. `next_cursor.offset` is 3.

#### 26. `consume-events/fetch-less-than-limit/`

- **Setup**: Produce 3 events. Run `fetch --consumer-id c1 --limit 5`.
- **Assert**: Returns 3 events. `has_more:false`.

#### 27. `consume-events/fetch-peek/`

- **Setup**: Produce 1 event. Run `fetch --consumer-id c1 --peek`.
- **Assert**: Returns 1 event. `previous_cursor == next_cursor`. Run `fetch --consumer-id c1` again → same event returned (cursor not advanced).

#### 28. `consume-events/fetch-advances-cursor/`

- **Setup**: Produce 3 events. First `fetch --limit 1`. Second `fetch --limit 1`.
- **Assert**: First returns event at offset 0. Second returns event at offset 1 (not offset 0).

#### 29. `consume-events/fetch-cross-partition/`

- **Setup**: Produce 1 event (in date partition A). Hook into a fake-opencode run that produces additional events on a second date partition. Fetch all.
- **Assert**: Single `fetch --limit 100` returns events spanning both partitions.

#### 30. `consume-events/fetch-empty/`

- **Setup**: No events. Run `fetch --consumer-id c1`.
- **Assert**: `events:[]`, `has_more:false`.

#### 31. `consume-events/fetch-missing-consumer-id/`

- **Setup**: `agent-hub fetch` without `--consumer-id`.
- **Assert**: ExitCode ≠ 0, error about consumer-id.

#### 32. `consume-events/replay-reset-cursor/`

- **Setup**: Produce 5 events. Fetch 3 (cursor at offset 3). Run `replay --consumer-id c1 --from <partition>:0`.
- **Assert**: Next `fetch --limit 5` returns 5 events (starts from offset 0 again).

#### 33. `consume-events/replay-bad-format/`

- **Setup**: `agent-hub replay --consumer-id c1 --from bogus`
- **Assert**: ExitCode ≠ 0, error about format.

---

### Group 4: Query Session (9 leaves)

Sessions created via `agent-hub notify` events. All use `AGENT_HUB_OPENCODE_RUNNER=fake-opencode`.

#### 34. `query-session/show-running/`

- **Setup**: Notify `agent.session.started` with `runner:"fake-opencode"`, `runner_session_id:"s_run"`. Run `session show --runner fake-opencode --session-id s_run`.
- **Assert**: ExitCode 0. JSON has `status:"running"`, `runner:"fake-opencode"`, `runner_session_id:"s_run"`, `last_event` with partition and offset.

#### 35. `query-session/show-completed/`

- **Setup**: Notify started + finished events. Run `session show`.
- **Assert**: `status:"completed"`.

#### 36. `query-session/show-failed/`

- **Setup**: Notify started + failed events. Run `session show`.
- **Assert**: `status:"failed"`.

#### 37. `query-session/show-non-existent/`

- **Setup**: No session exists. Run `session show --runner fake-opencode --session-id nosuch`.
- **Assert**: ExitCode ≠ 0, error about session not found.

#### 38. `query-session/show-missing-runner/`

- **Setup**: `agent-hub session show --session-id s1`
- **Assert**: ExitCode ≠ 0, error about `--runner`.

#### 39. `query-session/show-missing-session-id/`

- **Setup**: `agent-hub session show --runner fake-opencode`
- **Assert**: ExitCode ≠ 0, error about `--session-id`.

#### 40. `query-session/show-after-transition/`

- **Setup**: Notify started. `session show` → `status:"running"`. Notify finished. `session show` → `status:"completed"`.
- **Assert**: Both responses have correct status and the same `runner_session_id`.

#### 41. `query-session/list-all/`

- **Setup**: Create session for `fake-opencode` (via notify) and `codex` (via notify).
- **Assert**: `agent-hub sessions` returns JSON array with 2 entries. Each has `runner`, `runner_session_id`, `status`.

#### 42. `query-session/list-empty/`

- **Setup**: No sessions exist.
- **Assert**: `agent-hub sessions` returns `{"sessions":[]}`.

---

## How to Run

From the repo root:

```sh
doctest build agents/agent-hub/tests
doctest test -v agents/agent-hub/tests
```

All tests must pass with `AGENT_HUB_HOME` isolated per test (via `t.TempDir()`).

---

## Expected RED State

Before implementation, all 42 leaves must fail — the stub `Run` returns
`"error not implemented"` and all `Assert` blocks fail. Example failure:

```
--- FAIL: TestLeaf (0.00s)
    SETUP.md: error not implemented
```

No test should pass before implementation begins.
