# Scenario

**Feature**: attach to live real-grok session while run is blocking

```
background run --agent-runner grok-tty "say hi" → parse stderr id → attach connects
```

## Steps

1. Start background real-grok run with prompt `say hi`.
2. Parse session id from stderr.
3. `Run` probes attach via registry (WS + snapshot).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokTTYPrompt = "say hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "say hi"}
	startGrokTTYBackground(t, req)
	req.Mode = "attach-interactive-probe"
	return nil
}
```