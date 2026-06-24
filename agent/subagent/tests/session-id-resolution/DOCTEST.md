# Session ID Resolution Tests

Verify `resolveSessionID` behavior in `github.com/xhd2015/agent-pro/agent/subagent` when invoked via `Run(Status)`: flag, role env var, `CODEX_THREAD_ID` fallback, and `Config.AutoGenerateSessionID` policy.

## Version
0.0.2

# DSN (Domain Specific Notion)

**resolveSessionID** selects a session ID using fixed priority: `--session-id` flag, then `AGENT_PRO_SUBAGENT_<ROLE>_SESSION_ID`, then `CODEX_THREAD_ID`. When all sources miss, **Config.AutoGenerateSessionID** chooses the outcome: **Require** (default) returns an error with a retry hint containing a suggested `gen_*` ID; **AutoGenerate** calls `generateSessionID()` and proceeds with `source: generated`. **showStatus** then looks up the session directory and reports `session not found` when the ID has no on-disk session yet.

## Decision Tree

```
session-id-resolution/
├── DOCTEST.md
├── SETUP.md                           # Operation=session_id_resolution, runSessionIDResolution
├── from-env-var/                      # env var → session ID
├── flag-overrides-env/                # flag beats env var
├── codex-thread-id-fallback/          # CODEX_THREAD_ID fallback
├── no-session-id-error/               # Require policy → cannot detect session id
└── auto-generate-session-id/          # AutoGenerate policy → generated ID, session not found
```

## Test Index

| Leaf | Policy / Source | Expected |
|------|-----------------|----------|
| `from-env-var` | Env var | stderr references env session ID, session not found |
| `flag-overrides-env` | Flag over env | stderr references flag ID, not env ID |
| `codex-thread-id-fallback` | CODEX_THREAD_ID | stderr references codex thread ID |
| `no-session-id-error` | Require (default) | stderr contains cannot detect + gen_ retry hint |
| `auto-generate-session-id` | AutoGenerate | no cannot detect; session not found |

Total: **5 leaves**.

## How to Run

```sh
cd external/agent-pro-task-hub
doctest vet ./agent/subagent/tests/session-id-resolution/
doctest test -v ./agent/subagent/tests/session-id-resolution/
```

```go
import (
    "testing"
)

type Request struct {
    RoleName              string
    SessionID             string
    Env                   []string
    AutoGenerateSessionID bool
}

type Response struct {
    Stdout string
    Stderr string
    Err    error
}

func Run(t *testing.T, req *Request) (*Response, error) {
    return runSessionIDResolution(t, req)
}
```