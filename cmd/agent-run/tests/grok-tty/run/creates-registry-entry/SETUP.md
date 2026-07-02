# Scenario

**Feature**: registry JSON written while grok-tty run is active

```
background run (sleep 30 hook) → grok-tty-registry/<id>.json with listen_addr
```

## Steps

1. Start background `run` with long-running fake TUI.
2. Wait for stderr session id.
3. `Run` reads registry entry while process still alive.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokTTYCommand = fakeTUILongRun()
	req.GrokTTYPrompt = "hold"
	startGrokTTYBackground(t, req)
	req.Mode = "registry-while-running"
	return nil
}
```