# Crush Agent Tests

These doc-style tests verify the `agent/cli/crush` package — event conversion (`UnwrapEvent`), server client integration (`CrushServerClient`), and the `CrushAgent.Ask()` method in both subprocess and server modes.

## Decision Tree

```
tests/
├── basic-query/                        (Mode="",  subprocess Ask)
│   └── ASSERT: answer contains "paris"
├── session-resume/                     (Mode="",  subprocess Ask + resume)
│   └── ASSERT: answer references "french" or "capital"
│
└── server/                             (grouping: server-mode tests)
    │
    ├── convert/                        (grouping: Mode="convert")
    │   ├── message-event/              3-level message SSE → EventMessage
    │   │   └── ASSERT: type, role, session_id, parts
    │   ├── agent-error/                3-level agent_event SSE → EventAgentEvent
    │   │   └── ASSERT: type, error, run_id
    │   ├── run-complete/               3-level run_complete SSE → EventRunComplete
    │   │   └── ASSERT: session_id, run_id, text, error, cancelled
    │   ├── drop-unknown/               unknown outer type → nil (dropped)
    │   │   └── ASSERT: nil result, no error
    │   └── malformed-json/             invalid JSON → nil (graceful drop)
    │       └── ASSERT: nil result, no error
    │
    ├── server-client/                  (grouping: Mode="server-client")
    │   ├── health-check/               probe /v1/health → 200 OK
    │   │   └── ASSERT: status 200
    │   ├── auto-start/                 server missing → auto-starts
    │   │   └── ASSERT: health passes after start
    │   ├── create-workspace/           POST /v1/workspaces → workspace id
    │   │   └── ASSERT: non-empty id
    │   └── send-and-receive/           full SSE cycle: subscribe + send + read
    │       └── ASSERT: message events with text, then run_complete
    │
    └── session-persist/                (Mode="server-ask", resume session)
        └── ASSERT: second answer references first query context
```

## Test Leaves

| Leaf | Mode | Description |
|---|---|---|
| `basic-query` | `""` | Subprocess: capital of France → "paris" |
| `session-resume` | `""` | Subprocess: ask, resume, verify context |
| `server/convert/message-event` | `convert` | `UnwrapEvent`: message SSE → EventMessage with correct fields |
| `server/convert/agent-error` | `convert` | `UnwrapEvent`: agent_event SSE → EventAgentEvent with type=error |
| `server/convert/run-complete` | `convert` | `UnwrapEvent`: run_complete SSE → EventRunComplete |
| `server/convert/drop-unknown` | `convert` | `UnwrapEvent`: unknown type → nil (dropped) |
| `server/convert/malformed-json` | `convert` | `UnwrapEvent`: garbage input → nil (dropped) |
| `server/server-client/health-check` | `server-client` | `ensureServer`: probe /v1/health returns 200 |
| `server/server-client/auto-start` | `server-client` | `ensureServer`: auto-start when server not running |
| `server/server-client/create-workspace` | `server-client` | `createWorkspace`: POST returns valid id |
| `server/server-client/send-and-receive` | `server-client` | Full cycle: subscribe, send, read SSE events |
| `server/session-persist` | `server-ask` | Server-mode: Ask with session resume |

## How to Run

```sh
doctest test ./agent/cli/crush/tests/...
```

Integration tests (server-client leaves, session-persist) require a running crush server and the env var:

```sh
CRUSH_INTEGRATION_TEST=1 doctest test ./agent/cli/crush/tests/...
```
