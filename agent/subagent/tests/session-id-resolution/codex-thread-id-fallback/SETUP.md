# Scenario

**Feature**: CODEX_THREAD_ID fallback when flag and env miss

```
# CODEX_THREAD_ID is third priority source
CODEX_THREAD_ID=codex_thread_abc -> resolveSessionID -> showStatus
```

## Preconditions
- No `--session-id` flag and no `AGENT_PRO_SUBAGENT_TESTROLE_SESSION_ID` env var.
- `CODEX_THREAD_ID` env var is set.

## Steps
1. Leave `req.SessionID` empty.
2. Unset any `AGENT_PRO_SUBAGENT_*_SESSION_ID` env var.
3. Set `CODEX_THREAD_ID=codex_thread_abc` in `req.Env`.
4. Verify `CODEX_THREAD_ID` is used as fallback.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.SessionID = ""
    req.Env = append(req.Env,
        "CODEX_THREAD_ID=codex_thread_abc",
    )
    return nil
}
```
