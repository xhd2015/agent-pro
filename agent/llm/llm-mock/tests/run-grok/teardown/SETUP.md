# Scenario

**Feature**: `llm-mock run grok` must return promptly after grok exits (no multi-minute hang)

```
grok exits (e.g. /exit) -> orchestrator tears down mock + returns
must NOT block ~60s in waitAndMirrorSessions when session dir lacks events.jsonl
```

## Preconditions

- Uses fake `grok` on PATH (`GrokPathPrepend`), not `LLM_MOCK_RUN_GROK_COMMAND` hook,
  so `waitAndMirrorSessions` runs (real-grok code path).
- Leaves set short `ExecTimeout` (e.g. 5s) to detect stalls.

## Steps

1. Grouping `Setup` documents teardown contract.
2. Leaf `Setup` installs fake grok behavior and optional `--log-events`.
3. `Run` executes orchestrator with timeout.
4. `Assert` expects exit 0 without context timeout.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 5 * time.Second
	}
	return nil
}
```