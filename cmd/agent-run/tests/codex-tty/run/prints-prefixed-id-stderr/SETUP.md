# Scenario

**Feature**: session id printed to stderr with `codex-tty:` prefix

```
agent-run run --agent-runner codex-tty "hi" → stderr codex-tty: session-N; stdout lacks prefix
```

## Steps

1. Run with fake TUI respond script and prompt `hi`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CodexTTYCommand = fakeTUIRespondHi()
	req.Args = append(req.Args, "hi")
	return nil
}
```