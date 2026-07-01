# Scenario

**Feature**: Ask() surfaces a clear error when the claude binary path does not exist

```
# bogus AgentPath: ClaudeAgent tries to exec /nonexistent/claude
binary-not-found -> ClaudeAgent{AgentPath:"/nonexistent/claude"}.Ask(prompt)
ClaudeAgent -> exec /nonexistent/claude (fails)
ClaudeAgent <- error mentioning "claude"
```

## Preconditions
- This leaf does NOT require the `claude` binary (it never spawns a real one).
- The request sets an explicit, non-existent `AgentPath` so `Run` does not skip.

## Steps
1. Set `AgentPath` to a path that does not exist on disk.
2. Set a minimal prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentPath = "/nonexistent/claude"
	req.Prompt = "ping"
	return nil
}
```
