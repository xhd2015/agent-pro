# Scenario

**Feature**: attach to live real-codex session while run is blocking

```
background run --agent-runner codex-tty "say hi" → parse stderr id → attach connects
```

## Steps

1. Start background real-codex run with prompt `say hi`.
2. Parse session id from stderr.
3. `Run` probes attach via registry (WS + snapshot).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CodexTTYPrompt = "say hi"
	req.Args = []string{"run", "--agent-runner", "codex-tty", "say hi"}
	startCodexTTYBackground(t, req)
	req.Mode = "attach-interactive-probe"
	return nil
}
```