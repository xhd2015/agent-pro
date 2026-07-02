# Scenario

**Feature**: attach resolves hidden port from registry and completes WS handshake

```
background codex-tty run (sleep 30) → parse stderr id → attach probe via registry
```

## Steps

1. Start background run with long fake TUI.
2. Parse `codex-tty: session-N` from stderr.
3. `Run` probes attach through registry `listen_addr`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CodexTTYCommand = fakeTUILongRun()
	req.CodexTTYPrompt = "hold"
	startCodexTTYBackground(t, req)
	req.Mode = "attach-probe"
	return nil
}
```