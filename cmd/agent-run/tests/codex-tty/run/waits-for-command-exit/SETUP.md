# Scenario

**Feature**: `run` blocks until the PTY command exits

```
fake TUI exits after echo → agent-run run returns exit 0
```

## Steps

1. Run with fake TUI respond script; prompt `ping`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CodexTTYCommand = fakeTUIRespondHi()
	req.Args = append(req.Args, "ping")
	return nil
}
```