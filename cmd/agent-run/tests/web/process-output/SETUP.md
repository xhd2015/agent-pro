# Scenario

**Bug**: `agent-run web` must not print background agent human output on the server terminal

```
agent-run web (stderr: token warning + listen URL) -> POST session -> agentui.Run
# agent formatted lines must not appear on web process stdout/stderr
```

## Preconditions

- Web server runs in background with stderr/stdout captured incrementally (`webProcessStderr`, `webProcessStdout`).
- `fake-codex` on `PATH` for deterministic assistant completion.
- Human-readable agent events use `eventprint.FormatAgentEvent` (`💬`, `[done]`) on the run stdout writer.

## Steps

1. Leaf starts `agent-run web` and performs HTTP actions (create session, wait for finish).
2. `Run` may be a no-op GET; assertions read captured web process streams after the agent run completes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	return nil
}
```