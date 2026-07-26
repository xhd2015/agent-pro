# Scenario

**Feature**: empty AgentRunner omits `--agent-runner` in `Run`/`BuildArgs`

```
BuildArgs(AgentRunner="", session, prompt) -> no --agent-runner flag
```

## Preconditions

- Contrasts with InteractiveOpen which injects `grok-tty` when empty.

## Steps

1. Set session + prompt; leave AgentRunner empty; no open profile required.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-no-runner"
	req.Prompt = "no runner flag"
	req.AgentRunner = ""
	req.KeepTTY = true
	return nil
}
```
